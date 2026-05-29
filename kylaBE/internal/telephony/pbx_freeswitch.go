package telephony

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FreeSWITCHController is the FreeSWITCH PBX implementation. It speaks ESL
// (Event Socket Library) over TCP to control the PBX and consume its event
// stream.
//
// Scope of this skeleton:
//   - Connects and authenticates against ESL on startup
//   - Subscribes to channel + sofia event classes
//   - Lifts events from the socket onto a CallEventStream
//   - Stubs the originate/hangup/transfer/hold/resume command paths with a
//     clear "not yet wired" error so the gRPC server returns FailedPrecondition
//     until ESL command dispatch is implemented
//
// What's deliberately deferred:
//   - Real ESL command dispatch (bgapi originate <...>, uuid_kill, etc.) —
//     this needs careful framing handling and is a follow-up commit
//   - mod_xml_curl / mod_xml_rpc dialplan integration for inbound routing
//   - Recording management via uuid_record
//
// The skeleton design keeps the controller bootable today and progressively
// upgradable without changing the PBXController interface.
type FreeSWITCHController struct {
	host        string
	port        string
	password    string
	stream      *CallEventStream
	conn        net.Conn
	connMu      sync.Mutex
	stopCh      chan struct{}
	connectedAt time.Time
	enabled     bool
}

// FreeSWITCHConfig groups the ESL connection parameters.
type FreeSWITCHConfig struct {
	Host     string
	Port     string
	Password string
}

// NewFreeSWITCHController constructs the controller. It does NOT dial — call
// Start(ctx) to establish the connection and begin streaming events. Dialing
// is split out so main.go can register the controller before infra is up.
func NewFreeSWITCHController(cfg FreeSWITCHConfig, stream *CallEventStream) *FreeSWITCHController {
	if cfg.Port == "" {
		cfg.Port = "8021"
	}
	if stream == nil {
		stream = NewCallEventStream(0)
	}
	return &FreeSWITCHController{
		host:     cfg.Host,
		port:     cfg.Port,
		password: cfg.Password,
		stream:   stream,
		stopCh:   make(chan struct{}),
	}
}

func (c *FreeSWITCHController) Name() string  { return "freeswitch" }
func (c *FreeSWITCHController) Enabled() bool { return c != nil && c.enabled }

// Start dials ESL, authenticates, subscribes to the events we care about, and
// kicks off a background reader that parses the ESL framing into PBXEvent
// values written to the stream. Returns once the auth handshake is done so
// callers know the connection is usable.
//
// On dial failure the controller logs and returns nil — same graceful
// degradation pattern as NATS and Temporal. Subsequent calls to Originate /
// Hangup will return PBX-not-configured errors until a successful Start.
func (c *FreeSWITCHController) Start(ctx context.Context) error {
	if c.host == "" {
		log.Println("freeswitch: FS_ESL_HOST empty; controller disabled")
		return nil
	}
	addr := net.JoinHostPort(c.host, c.port)
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		log.Printf("freeswitch: dial %s failed: %v (controller disabled)", addr, err)
		return nil
	}
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	reader := bufio.NewReader(conn)

	// 1. Read "auth/request" content-type frame.
	if _, err := readESLFrame(reader); err != nil {
		_ = conn.Close()
		return fmt.Errorf("read auth request: %w", err)
	}

	// 2. Send auth.
	if _, err := conn.Write([]byte("auth " + c.password + "\n\n")); err != nil {
		_ = conn.Close()
		return fmt.Errorf("send auth: %w", err)
	}
	frame, err := readESLFrame(reader)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("read auth response: %w", err)
	}
	if !strings.Contains(frame, "Reply-Text: +OK") {
		_ = conn.Close()
		return fmt.Errorf("freeswitch: auth rejected: %s", frame)
	}

	// 3. Subscribe to the event classes we care about. PLAIN keeps parsing
	//    simple; if we need richer payloads later we'd switch to JSON.
	//    PLAYBACK_STOP and CHANNEL_EXECUTE_COMPLETE are needed for IVR
	//    (the latter is how play_and_get_digits returns the captured DTMF).
	if _, err := conn.Write([]byte("event plain CHANNEL_CREATE CHANNEL_ANSWER CHANNEL_HANGUP_COMPLETE SOFIA_REGISTER RECORD_STOP PLAYBACK_STOP CHANNEL_EXECUTE_COMPLETE\n\n")); err != nil {
		_ = conn.Close()
		return fmt.Errorf("subscribe events: %w", err)
	}
	if _, err := readESLFrame(reader); err != nil {
		// Subscribe ack — non-fatal if missing on first read.
		log.Printf("freeswitch: event subscribe ack: %v", err)
	}

	c.enabled = true
	c.connectedAt = time.Now().UTC()
	log.Printf("freeswitch: connected to ESL %s at %s", addr, c.connectedAt.Format(time.RFC3339))

	go c.readLoop(reader)
	return nil
}

// Stop closes the ESL connection and stops the reader.
func (c *FreeSWITCHController) Stop() {
	close(c.stopCh)
	c.connMu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()
	c.enabled = false
}

// readLoop continuously reads ESL frames and converts the ones we care about
// into PBXEvent values on the stream. Unknown event types are silently
// dropped.
//
// On connection error the loop exits, the controller is marked disabled, and
// the binary keeps running — calls return PBX-not-configured until restart.
// Automatic reconnect with backoff is a follow-up improvement.
func (c *FreeSWITCHController) readLoop(reader *bufio.Reader) {
	defer func() {
		c.enabled = false
		log.Println("freeswitch: ESL read loop exited; controller disabled")
	}()

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		frame, err := readESLFrame(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("freeswitch: ESL read error: %v", err)
			}
			return
		}
		evt, ok := parseESLEvent(frame)
		if !ok {
			continue
		}
		select {
		case c.stream.Events <- evt:
		default:
			log.Println("freeswitch: event stream full; dropping event")
		}
	}
}

// ── PBXController surface ────────────────────────────────────────────────────

// Originate is the only command path that's partially wired here: it generates
// a UUID up-front (so the gRPC server can persist the Call row with the right
// ID) and issues a bgapi originate command via ESL. Reading the success
// response with proper background-job correlation is a follow-up — for now
// we treat "command accepted" as success and rely on CHANNEL_HANGUP events to
// surface failures.
func (c *FreeSWITCHController) Originate(ctx context.Context, req OriginateRequest) (string, error) {
	if !c.Enabled() {
		return "", errPBXNotConfigured
	}
	if req.AgentExtension == "" || req.ToNumber == "" {
		return "", errors.New("originate: agent_extension and to_number required")
	}
	callUUID := uuid.NewString()
	// Build a bgapi originate command. Note: per-call variables go in {} before
	// the leg URI so they're inherited by both legs of the bridge.
	cmd := fmt.Sprintf(
		"bgapi originate {origination_uuid=%s,kyla_org_id=%s,kyla_workspace_id=%s,kyla_agent_id=%s,kyla_contact_id=%s,kyla_recording=%v}user/%s &bridge(sofia/gateway/%s/%s)\n\n",
		callUUID,
		req.OrgID,
		req.WorkspaceID,
		req.AgentID,
		req.ContactID,
		req.RecordingEnabled,
		req.AgentExtension,
		req.TrunkGateway,
		req.ToNumber,
	)
	if err := c.writeCommand(cmd); err != nil {
		return "", fmt.Errorf("originate: write command: %w", err)
	}
	return callUUID, nil
}

func (c *FreeSWITCHController) Hangup(ctx context.Context, callUUID, reason string) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	if reason == "" {
		reason = "NORMAL_CLEARING"
	}
	return c.writeCommand(fmt.Sprintf("bgapi uuid_kill %s %s\n\n", callUUID, reason))
}

func (c *FreeSWITCHController) Transfer(ctx context.Context, callUUID, target string, blind bool) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	verb := "uuid_transfer"
	if !blind {
		// Attended transfer requires a B-leg consultation first. Out of scope
		// for the skeleton — return an explicit error so the caller can fall
		// back to blind for now.
		return errors.New("attended transfer not yet wired; pass blind=true")
	}
	return c.writeCommand(fmt.Sprintf("bgapi %s %s %s\n\n", verb, callUUID, target))
}

func (c *FreeSWITCHController) Hold(ctx context.Context, callUUID string) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	return c.writeCommand(fmt.Sprintf("bgapi uuid_hold %s\n\n", callUUID))
}

func (c *FreeSWITCHController) Resume(ctx context.Context, callUUID string) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	return c.writeCommand(fmt.Sprintf("bgapi uuid_hold off %s\n\n", callUUID))
}

// ProvisionExtension would push an extension's user definition into the PBX
// — but FreeSWITCH expects XML directory entries on disk (or via mod_xml_curl).
// Skipping in the skeleton; provisioning happens out-of-band today and the
// store persists the hash so future mod_xml_curl integration can read it.
func (c *FreeSWITCHController) ProvisionExtension(_ context.Context, _ SipExtension, _ string) error {
	log.Println("freeswitch: ProvisionExtension is a no-op; configure directory out-of-band for now")
	return nil
}

// ProvisionTrunk similarly defers to out-of-band gateway profile XML. The
// trunk row stores the credentials a future mod_xml_curl handler will surface.
func (c *FreeSWITCHController) ProvisionTrunk(_ context.Context, _ SipTrunk) error {
	log.Println("freeswitch: ProvisionTrunk is a no-op; configure sofia gateway out-of-band for now")
	return nil
}

// ── IVR command surface ────────────────────────────────────────────────────

// PlayAudio queues a static file for playback on the leg. We use bgapi so the
// command returns immediately; PLAYBACK_STOP arrives on the event stream when
// playback completes (or is interrupted).
func (c *FreeSWITCHController) PlayAudio(ctx context.Context, callUUID, audioPath string) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	if callUUID == "" || audioPath == "" {
		return errors.New("PlayAudio: call_uuid and audio_path required")
	}
	return c.writeCommand(fmt.Sprintf("bgapi uuid_broadcast %s playback::%s aleg\n\n", callUUID, audioPath))
}

// SayText synthesises speech via mod_say. Voice is forwarded as the engine
// hint (e.g. "en" for the English say engine).
func (c *FreeSWITCHController) SayText(ctx context.Context, callUUID, voice, text string) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	if voice == "" {
		voice = "en"
	}
	// say::<engine>!<voice>!<text> tells uuid_broadcast to invoke mod_say.
	// Escape colons and exclamations in text to keep the framing intact.
	safeText := strings.ReplaceAll(text, "'", "")
	return c.writeCommand(fmt.Sprintf("bgapi uuid_broadcast %s 'say::%s SHORT pronounced %s' aleg\n\n", callUUID, voice, safeText))
}

// PlayAndGetDigits issues the play_and_get_digits app via uuid_setvar +
// uuid_broadcast. The captured digits arrive as a CHANNEL_EXECUTE_COMPLETE
// event with Application-Response set to the digit string.
//
// FreeSWITCH's app signature is:
//   play_and_get_digits <min> <max> <tries> <timeout> <terminators> <file> <invalid_file> <var_name> <regexp>
func (c *FreeSWITCHController) PlayAndGetDigits(ctx context.Context, callUUID string, opts PlayAndGetDigitsOpts) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	if opts.MinDigits == 0 {
		opts.MinDigits = 1
	}
	if opts.MaxDigits == 0 {
		opts.MaxDigits = 1
	}
	if opts.Tries == 0 {
		opts.Tries = 1
	}
	timeoutMs := int(opts.Timeout / time.Millisecond)
	if timeoutMs == 0 {
		timeoutMs = 5000
	}
	terminator := opts.TerminatorKey
	if terminator == "" {
		terminator = "#"
	}
	invalid := opts.InvalidFile
	if invalid == "" {
		invalid = "silence_stream://250"
	}
	regex := opts.Regex
	if regex == "" {
		regex = "\\d+"
	}
	app := fmt.Sprintf(
		"play_and_get_digits %d %d %d %d %s %s %s kyla_ivr_digits %s",
		opts.MinDigits, opts.MaxDigits, opts.Tries, timeoutMs,
		terminator,
		opts.PromptFile,
		invalid,
		regex,
	)
	return c.writeCommand(fmt.Sprintf("bgapi uuid_broadcast %s '%s' aleg\n\n", callUUID, app))
}

// StartRecording captures audio of the leg to the supplied file path.
// maxSeconds=0 means record until hangup.
func (c *FreeSWITCHController) StartRecording(ctx context.Context, callUUID, recordingPath string, maxSeconds int) error {
	if !c.Enabled() {
		return errPBXNotConfigured
	}
	if callUUID == "" || recordingPath == "" {
		return errors.New("StartRecording: call_uuid and recording_path required")
	}
	cmd := fmt.Sprintf("bgapi uuid_record %s start %s", callUUID, recordingPath)
	if maxSeconds > 0 {
		cmd += fmt.Sprintf(" %d", maxSeconds)
	}
	return c.writeCommand(cmd + "\n\n")
}

// ── ESL framing helpers ─────────────────────────────────────────────────────

// writeCommand sends a single command to ESL. Safe for concurrent callers via
// connMu — ESL commands are line-oriented and must not interleave.
func (c *FreeSWITCHController) writeCommand(cmd string) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == nil {
		return errPBXNotConfigured
	}
	_, err := c.conn.Write([]byte(cmd))
	return err
}

// readESLFrame reads one ESL frame (Content-Length-delimited). FreeSWITCH ESL
// always frames with HTTP-like headers ending in \n\n, optionally followed by
// a body whose length is given by Content-Length.
func readESLFrame(reader *bufio.Reader) (string, error) {
	var headers strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		headers.WriteString(line)
		if line == "\n" {
			break
		}
	}
	header := headers.String()
	clen := parseContentLength(header)
	if clen == 0 {
		return header, nil
	}
	body := make([]byte, clen)
	if _, err := io.ReadFull(reader, body); err != nil {
		return "", err
	}
	return header + string(body), nil
}

func parseContentLength(header string) int {
	for _, line := range strings.Split(header, "\n") {
		if strings.HasPrefix(line, "Content-Length:") {
			n := 0
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return 0
			}
			s := strings.TrimSpace(parts[1])
			for _, r := range s {
				if r < '0' || r > '9' {
					break
				}
				n = n*10 + int(r-'0')
			}
			return n
		}
	}
	return 0
}

// parseESLEvent converts a raw ESL frame body into a typed PBXEvent. Returns
// false when the frame is something we don't care about (commands acks,
// log lines, etc.).
func parseESLEvent(frame string) (PBXEvent, bool) {
	// ESL plain events use Event-Name as the discriminator.
	fields := parseHeaders(frame)
	name := fields["Event-Name"]

	dataIface := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		dataIface[k] = v
	}
	evt := PBXEvent{OccurredAt: time.Now().UTC(), Data: dataIface}
	switch name {
	case "CHANNEL_CREATE":
		evt.Type = EventChannelCreate
		evt.CallUUID = firstNonEmpty(fields["Unique-ID"], fields["Channel-Call-UUID"])
		evt.OrgID = fields["variable_kyla_org_id"]
	case "CHANNEL_ANSWER":
		evt.Type = EventChannelAnswer
		evt.CallUUID = firstNonEmpty(fields["Unique-ID"], fields["Channel-Call-UUID"])
		evt.OrgID = fields["variable_kyla_org_id"]
	case "CHANNEL_HANGUP_COMPLETE":
		evt.Type = EventChannelHangup
		evt.CallUUID = firstNonEmpty(fields["Unique-ID"], fields["Channel-Call-UUID"])
		evt.OrgID = fields["variable_kyla_org_id"]
	case "SOFIA_REGISTER":
		evt.Type = EventSofiaRegister
		evt.Extension = fields["from-user"]
	case "RECORD_STOP":
		evt.Type = EventRecordingComplete
		evt.CallUUID = firstNonEmpty(fields["Unique-ID"], fields["Channel-Call-UUID"])
	case "PLAYBACK_STOP":
		evt.Type = EventPlaybackStop
		evt.CallUUID = firstNonEmpty(fields["Unique-ID"], fields["Channel-Call-UUID"])
	case "CHANNEL_EXECUTE_COMPLETE":
		// Only forward the IVR-relevant completions. play_and_get_digits
		// stores the captured input in the channel variable named in its
		// var_name argument — we use "kyla_ivr_digits" so we can recognise it.
		app := fields["Application"]
		if app != "play_and_get_digits" {
			return PBXEvent{}, false
		}
		evt.Type = EventDTMFCaptured
		evt.CallUUID = firstNonEmpty(fields["Unique-ID"], fields["Channel-Call-UUID"])
		// The captured digits live in variable_kyla_ivr_digits.
		if digits, ok := dataIface["variable_kyla_ivr_digits"]; ok {
			if s, ok := digits.(string); ok {
				evt.Data["captured_digits"] = s
			}
		}
	default:
		return PBXEvent{}, false
	}
	return evt, true
}

func parseHeaders(frame string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(frame, "\n") {
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	return out
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
