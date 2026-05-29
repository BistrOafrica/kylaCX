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

	// Transfer moves a leg to a different destination. blind=true is a
	// release-and-redirect; blind=false is attended (consultation) transfer.
	Transfer(ctx context.Context, callUUID, target string, blind bool) error

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
