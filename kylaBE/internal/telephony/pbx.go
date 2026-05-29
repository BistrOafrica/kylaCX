package telephony

import (
	"context"
	"time"
)

// PBXController is the provider-agnostic control surface over the SIP PBX.
// The gRPC server depends only on this interface, so swapping FreeSWITCH for
// Asterisk (or running in cloud-provider-only mode) is a registration change
// in main.go — not a server rewrite.
type PBXController interface {
	// Name returns "freeswitch", "asterisk", "noop", etc. — used in logs.
	Name() string

	// Enabled reports whether the controller is connected and ready. Callers
	// short-circuit when false so a missing PBX returns FailedPrecondition
	// rather than crashing.
	Enabled() bool

	// Originate bridges an agent's extension to an external number via the
	// supplied trunk. Returns the PBX-assigned call UUID — the gRPC server
	// persists a Call row keyed by this UUID before returning.
	Originate(ctx context.Context, req OriginateRequest) (callUUID string, err error)

	// Hangup terminates an in-flight call. reason is propagated to the PBX
	// (e.g. "NORMAL_CLEARING").
	Hangup(ctx context.Context, callUUID, reason string) error

	// Transfer moves a leg to a different destination.
	//   blind=true   release-and-redirect: A → C, B drops out instantly.
	//                Returns ("", nil) on success.
	//   blind=false  attended (consultation): place A on hold, originate a
	//                B → C consultation leg. Returns the consultation leg's
	//                PBX UUID. The caller then either CompleteTransfer (bridge
	//                A ↔ C, kill B's leg to C) or Hangup the consultation leg
	//                to abort and resume A.
	Transfer(ctx context.Context, callUUID, target string, blind bool) (consultationUUID string, err error)

	// CompleteTransfer finalises an attended transfer started by
	// Transfer(blind=false). callerUUID is the original A leg; consultationUUID
	// is the UUID returned by the consultation Transfer call. The call bridges
	// A ↔ C and tears down the agent's consultation leg.
	CompleteTransfer(ctx context.Context, callerUUID, consultationUUID string) error

	// Hold/Resume park the leg's media without ending the signalling.
	Hold(ctx context.Context, callUUID string) error
	Resume(ctx context.Context, callUUID string) error

	// ProvisionExtension installs an extension in the PBX so the user's
	// softphone can register. plaintextPassword is given here exactly once;
	// the store persists only a hash.
	ProvisionExtension(ctx context.Context, ext SipExtension, plaintextPassword string) error

	// ProvisionTrunk registers an outbound gateway in the PBX from the trunk
	// row. Idempotent — repeated calls update an existing gateway profile.
	ProvisionTrunk(ctx context.Context, trunk SipTrunk) error

	// ── IVR command surface ──────────────────────────────────────────────────
	// These commands are issued by the IVR executor while a call is active.
	// All return an error if the PBX command socket isn't connected (NoopPBX
	// returns errPBXNotConfigured uniformly).

	// PlayAudio plays a static audio file on the leg. Returns once the
	// playback completes (FreeSWITCH's `uuid_broadcast` model is async; this
	// just queues the command — completion arrives as a PLAYBACK_STOP event).
	PlayAudio(ctx context.Context, callUUID, audioPath string) error

	// SayText synthesises text via the configured TTS engine.
	SayText(ctx context.Context, callUUID, voice, text string) error

	// PlayAndGetDigits is the workhorse of menu navigation. Plays a prompt
	// (file path), then collects up to maxDigits DTMF presses with a timeout.
	// The captured digits arrive as a DTMF_GET event correlated by callUUID.
	PlayAndGetDigits(ctx context.Context, callUUID string, opts PlayAndGetDigitsOpts) error

	// StartRecording begins recording the audio of the call to the supplied
	// path. Stop arrives via RECORD_STOP or call hangup.
	StartRecording(ctx context.Context, callUUID, recordingPath string, maxSeconds int) error
}

// PlayAndGetDigitsOpts groups the parameters for the menu primitive.
// Mirrors FreeSWITCH's play_and_get_digits application closely.
type PlayAndGetDigitsOpts struct {
	PromptFile   string        // file path to play; empty means no prompt
	MinDigits    int
	MaxDigits    int
	Tries        int           // how many times to re-prompt on no input
	Timeout      time.Duration // per-prompt timeout
	TerminatorKey string       // typically "#"
	InvalidFile   string       // played when input doesn't match regex (optional)
	Regex         string       // validates the captured digit string (e.g. "^[0-9]$")
}

// OriginateRequest is the controller-level request shape for outbound dials.
// Separate from the gRPC request type so the PBX layer is free of pb imports.
type OriginateRequest struct {
	AgentID          string
	AgentExtension   string
	ToNumber         string
	FromNumber       string
	TrunkGateway     string
	OrgID            string
	WorkspaceID      string
	ContactID        string
	RecordingEnabled bool
}

// CallEventStream is the channel-based stream of PBX events. The ESL controller
// publishes onto Events; the bridge subscribes and updates the Call store +
// emits NATS events for downstream consumers (VoiceCallBridge, analytics).
//
// Each PBXController is responsible for publishing into the stream; the bridge
// is the single consumer.
type CallEventStream struct {
	Events chan PBXEvent
}

// PBXEventType enumerates the subset of PBX events we currently care about.
// More types will be added as features land (DTMF, recording state, queue
// position updates, etc.).
type PBXEventType string

const (
	EventChannelCreate     PBXEventType = "channel_create"
	EventChannelAnswer     PBXEventType = "channel_answer"
	EventChannelHangup     PBXEventType = "channel_hangup"
	EventSofiaRegister     PBXEventType = "sofia_register"
	EventRecordingComplete PBXEventType = "recording_complete"
	EventPlaybackStop      PBXEventType = "playback_stop"   // a play_audio / say finished
	EventDTMFCaptured      PBXEventType = "dtmf_captured"    // play_and_get_digits returned input
)

// PBXEvent is one event lifted from the PBX event stream. CallUUID is the
// PBX-assigned UUID; everything else is event-specific.
type PBXEvent struct {
	Type       PBXEventType           `json:"type"`
	CallUUID   string                 `json:"call_uuid,omitempty"`
	Extension  string                 `json:"extension,omitempty"`
	OrgID      string                 `json:"org_id,omitempty"`
	OccurredAt time.Time              `json:"occurred_at"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// NewCallEventStream constructs a buffered event stream. Buffer size is sized
// for ~1 minute of normal-load events at 100 calls/minute; bursts beyond that
// block the producer, which is desirable backpressure.
func NewCallEventStream(buffer int) *CallEventStream {
	if buffer <= 0 {
		buffer = 1024
	}
	return &CallEventStream{Events: make(chan PBXEvent, buffer)}
}
