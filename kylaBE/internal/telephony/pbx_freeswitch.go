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

	// Job correlation: bgapi commands return synchronously with a Job-UUID,
	// and the actual result arrives later as a BACKGROUND_JOB event keyed by
	// that UUID. pendingJobs maps Job-UUID → result channel so Originate can
	// await the +OK / -ERR response synchronously without blocking other
	// commands. Buffered channel (size 1) so the event reader never blocks
	// even if the caller has given up.
	jobsMu      sync.Mutex
	pendingJobs map[string]chan bgapiResult
}

// bgapiResult is the structured outcome of a bgapi command. Body is the raw
// FreeSWITCH response (e.g. "+OK <uuid>\n" or "-ERR <reason>\n").
type bgapiResult struct {
	OK   bool
	Body string
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
		host:        cfg.Host,
		port:        cfg.Port,
		password:    cfg.Password,
		stream:      stream,
		stopCh:      make(chan struct{}),
		pendingJobs: map[string]chan bgapiResult{},
	}
}

func (c *FreeSWITCHController) Name() string  { return "freeswitch" }
func (c *FreeSWITCHController) Enabled() bool { return c != nil && c.enabled }

// Start launches the ESL supervisor goroutine. The supervisor dials ESL,
// authenticates, subscribes to events, runs the read loop, and reconnects on
// disconnect with exponential backoff (1s → 30s ceiling). Returns nil
// immediately — the caller doesn't block waiting for the first connection.
//
// Graceful degradation matches the rest of the stack: empty host → no-op.
// Persistent reconnect failures keep the controller disabled but never crash
// the binary; gRPC call-control RPCs return FailedPrecondition until a
// successful reconnect.
func (c *FreeSWITCHController) Start(ctx context.Context) error {
	if c.host == "" {
		log.Println("freeswitch: FS_ESL_HOST empty; controller disabled")
		return nil
	}
	go c.supervise(ctx)
	return nil
}

// supervise is the reconnect loop. Runs until ctx is cancelled or Stop is
// called. Each iteration attempts a connection; on failure it sleeps with
// exponential backoff before retrying. On clean disconnect it reconnects
// immediately (small jitter).
func (c *FreeSWITCHController) supervise(ctx context.Context) {
	const (
		minBackoff = 1 * time.Second
		maxBackoff = 30 * time.Second
	)
	backoff := minBackoff
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
		}
		if err := c.connectOnce(ctx); err != nil {
			log.Printf("freeswitch: connection attempt failed: %v (retry in %s)", err, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			case <-c.stopCh:
				return
			}
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		// connectOnce returned cleanly — the read loop terminated. Reset
		// backoff and reconnect immediately so a transient ESL bounce is
		// invisible to the operator.
		backoff = minBackoff
		log.Println("freeswitch: ESL connection ended; reconnecting")
	}
}

// connectOnce performs a single dial+handshake+readLoop cycle. Returns when
// the read loop exits (clean EOF or read error). The caller (supervise) is
// responsible for backoff + retry.
func (c *FreeSWITCHController) connectOnce(ctx context.Context) error {
	addr := net.JoinHostPort(c.host, c.port)
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
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
		return fmt.Errorf("auth rejected: %s", frame)
	}

	// 3. Subscribe to the event classes we care about. BACKGROUND_JOB is
	//    needed for bgapi command-result correlation; everything else is
	//    the call lifecycle + IVR-relevant events.
	if _, err := conn.Write([]byte("event plain CHANNEL_CREATE CHANNEL_ANSWER CHANNEL_HANGUP_COMPLETE SOFIA_REGISTER RECORD_STOP PLAYBACK_STOP CHANNEL_EXECUTE_COMPLETE BACKGROUND_JOB\n\n")); err != nil {
		_ = conn.Close()
		return fmt.Errorf("subscribe events: %w", err)
	}
	if _, err := readESLFrame(reader); err != nil {
		// Subscribe ack — non-fatal if missing on first read.
		log.Printf("freeswitch: event subscribe ack: %v", err)
	}

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
	c.enabled = true
	c.connectedAt = time.Now().UTC()
	log.Printf("freeswitch: connected to ESL %s at %s", addr, c.connectedAt.Format(time.RFC3339))

	c.readLoop(reader)

	// Read loop exited. Mark disabled + drop any pending job waiters so they
	// see a clean failure rather than blocking forever.
	c.connMu.Lock()
	c.conn = nil
	c.enabled = false
	c.connMu.Unlock()
	c.drainPendingJobs("esl_disconnected")
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

// readLoop continuously reads ESL frames. BACKGROUND_JOB frames are routed
// to any waiting pendingJobs channel for bgapi result correlation; other
// frames are converted to PBXEvent values and pushed to the event stream.
// Unknown frames are silently dropped.
func (c *FreeSWITCHController) readLoop(reader *bufio.Reader) {
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

		// BACKGROUND_JOB intercept: the frame carries Job-UUID + Body, and we
		// route it to the waiting channel rather than the public event stream.
		if strings.Contains(frame, "Event-Name: BACKGROUND_JOB") {
			c.handleBackgroundJob(frame)
			continue
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

// handleBackgroundJob extracts the Job-UUID from a BACKGROUND_JOB frame and
// delivers the body to the corresponding pending bgapi caller (if any).
func (c *FreeSWITCHController) handleBackgroundJob(frame string) {
	fields := parseHeaders(frame)
	jobUUID := fields["Job-UUID"]
	if jobUUID == "" {
		return
	}
	// The body sits after the headers' blank-line separator. ESL frames us
	// the full text in `frame` — extract everything after the first \n\n.
	body := ""
	if idx := strings.Index(frame, "\n\n"); idx >= 0 {
		body = strings.TrimSpace(frame[idx+2:])
	}
	c.jobsMu.Lock()
	ch, ok := c.pendingJobs[jobUUID]
	delete(c.pendingJobs, jobUUID)
	c.jobsMu.Unlock()
	if !ok {
		return
	}
	res := bgapiResult{Body: body, OK: strings.HasPrefix(body, "+OK")}
	select {
	case ch <- res:
	default:
		// Buffered channel; this should never block.
	}
}

// drainPendingJobs is called when the ESL connection drops. Surfaces a
// consistent failure to every caller currently blocked waiting for a bgapi
// result, rather than leaving them hanging until timeout.
func (c *FreeSWITCHController) drainPendingJobs(reason string) {
	c.jobsMu.Lock()
	pending := c.pendingJobs
	c.pendingJobs = map[string]chan bgapiResult{}
	c.jobsMu.Unlock()
	for _, ch := range pending {
		select {
		case ch <- bgapiResult{OK: false, Body: "-ERR " + reason}:
		default:
		}
	}
}

// runBgapi issues a bgapi command and waits for the BACKGROUND_JOB result.
// Returns the result body (e.g. "+OK <call-uuid>" or "-ERR USER_BUSY") so
// callers can extract whatever they need.
//
// The caller's context controls cancellation — a cancelled ctx unsubscribes
// the waiter and returns ctx.Err(); the eventual BACKGROUND_JOB is then
// dropped silently. Default safety timeout: 30s if the caller didn't set one.
func (c *FreeSWITCHController) runBgapi(ctx context.Context, command string) (bgapiResult, error) {
	if !c.Enabled() {
		return bgapiResult{}, errPBXNotConfigured
	}
	jobUUID := uuid.NewString()
	ch := make(chan bgapiResult, 1)

	c.jobsMu.Lock()
	c.pendingJobs[jobUUID] = ch
	c.jobsMu.Unlock()

	// "bgapi" + space + command + Job-UUID header + double newline.
	wire := fmt.Sprintf("bgapi %s\nJob-UUID: %s\n\n", command, jobUUID)
	if err := c.writeCommand(wire); err != nil {
		c.jobsMu.Lock()
		delete(c.pendingJobs, jobUUID)
		c.jobsMu.Unlock()
		return bgapiResult{}, fmt.Errorf("bgapi write: %w", err)
	}

	// Apply a 30s safety deadline when the caller didn't.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	select {
	case res := <-ch:
		return res, nil
	case <-ctx.Done():
		c.jobsMu.Lock()
		delete(c.pendingJobs, jobUUID)
		c.jobsMu.Unlock()
		return bgapiResult{}, ctx.Err()
	}
}

// ── PBXController surface ────────────────────────────────────────────────────

// Originate dials the supplied number via the agent's extension and the
// configured trunk. Generates the call UUID up-front so the gRPC server can
// persist the Call row keyed by the same UUID the PBX uses.
//
// Awaits the BACKGROUND_JOB result via runBgapi — returning the call UUID
// only when the PBX confirms the originate was accepted. Permanent failures
// (USER_BUSY, NO_ROUTE_DESTINATION, etc.) propagate as errors so the caller
// learns about them immediately rather than discovering them via the
// subsequent CHANNEL_HANGUP event.
func (c *FreeSWITCHController) Originate(ctx context.Context, req OriginateRequest) (string, error) {
	if !c.Enabled() {
		return "", errPBXNotConfigured
	}
	if req.AgentExtension == "" || req.ToNumber == "" {
		return "", errors.New("originate: agent_extension and to_number required")
	}
	callUUID := uuid.NewString()
	// Per-call variables go in {} before the leg URI so they're inherited by
	// both legs of the bridge.
	command := fmt.Sprintf(
		"originate {origination_uuid=%s,kyla_org_id=%s,kyla_workspace_id=%s,kyla_agent_id=%s,kyla_contact_id=%s,kyla_recording=%v}user/%s &bridge(sofia/gateway/%s/%s)",
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

	res, err := c.runBgapi(ctx, command)
	if err != nil {
		return "", fmt.Errorf("originate: %w", err)
	}
	if !res.OK {
		return "", fmt.Errorf("originate rejected by PBX: %s", res.Body)
	}
	return callUUID, nil
}

func (c *FreeSWITCHController) Hangup(ctx context.Context, callUUID, reason string) error {
	if reason == "" {
		reason = "NORMAL_CLEARING"
	}
	res, err := c.runBgapi(ctx, fmt.Sprintf("uuid_kill %s %s", callUUID, reason))
	if err != nil {
		return fmt.Errorf("hangup: %w", err)
	}
	if !res.OK {
		return fmt.Errorf("hangup rejected: %s", res.Body)
	}
	return nil
}

func (c *FreeSWITCHController) Transfer(ctx context.Context, callUUID, target string, blind bool) (string, error) {
	if blind {
		res, err := c.runBgapi(ctx, fmt.Sprintf("uuid_transfer %s %s", callUUID, target))
		if err != nil {
			return "", fmt.Errorf("transfer: %w", err)
		}
		if !res.OK {
			return "", fmt.Errorf("transfer rejected: %s", res.Body)
		}
		return "", nil
	}

	// Attended transfer.
	//   1. Park A (place on hold). Music-on-hold from the leg's sofia profile
	//      plays automatically while ESL holds the bridge open.
	//   2. Originate a B → C consultation leg. We generate the UUID up-front
	//      so the caller can correlate CompleteTransfer / Hangup back to the
	//      right leg without parsing the originate response.
	//
	// If the consultation originate fails we resume the A leg so the caller
	// isn't left on hold indefinitely.
	if err := c.Hold(ctx, callUUID); err != nil {
		return "", fmt.Errorf("attended transfer hold: %w", err)
	}
	consultUUID := uuid.NewString()
	// `&park` parks the consultation leg server-side once it's answered so the
	// agent can talk to C before the bridge is finalised. The kyla_attended_a
	// variable carries the original A-leg UUID so CompleteTransfer can find
	// both ends without a separate lookup.
	cmd := fmt.Sprintf(
		"originate {origination_uuid=%s,kyla_attended_a=%s}user/%s &park",
		consultUUID, callUUID, target,
	)
	res, err := c.runBgapi(ctx, cmd)
	if err != nil {
		_ = c.Resume(ctx, callUUID) // best-effort un-hold
		return "", fmt.Errorf("attended consultation: %w", err)
	}
	if !res.OK {
		_ = c.Resume(ctx, callUUID)
		return "", fmt.Errorf("attended consultation rejected: %s", res.Body)
	}
	return consultUUID, nil
}

// CompleteTransfer bridges the original A leg to the consultation leg and
// kills the operator's consultation leg. The PBX guarantees the bridge
// completes atomically — A and C end up on the same RTP session.
func (c *FreeSWITCHController) CompleteTransfer(ctx context.Context, callerUUID, consultationUUID string) error {
	if callerUUID == "" || consultationUUID == "" {
		return errors.New("complete_transfer: callerUUID and consultationUUID required")
	}
	res, err := c.runBgapi(ctx, fmt.Sprintf("uuid_bridge %s %s", callerUUID, consultationUUID))
	if err != nil {
		return fmt.Errorf("complete_transfer bridge: %w", err)
	}
	if !res.OK {
		return fmt.Errorf("complete_transfer rejected: %s", res.Body)
	}
	return nil
}

func (c *FreeSWITCHController) Hold(ctx context.Context, callUUID string) error {
	res, err := c.runBgapi(ctx, fmt.Sprintf("uuid_hold %s", callUUID))
	if err != nil {
		return fmt.Errorf("hold: %w", err)
	}
	if !res.OK {
		return fmt.Errorf("hold rejected: %s", res.Body)
	}
	return nil
}

func (c *FreeSWITCHController) Resume(ctx context.Context, callUUID string) error {
	res, err := c.runBgapi(ctx, fmt.Sprintf("uuid_hold off %s", callUUID))
	if err != nil {
		return fmt.Errorf("resume: %w", err)
	}
	if !res.OK {
		return fmt.Errorf("resume rejected: %s", res.Body)
	}
	return nil
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
