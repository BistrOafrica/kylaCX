package telephony

import (
	"context"
	"errors"
)

// NoopPBX is the fallback controller used when no PBX is configured. Every
// method returns a clear "PBX not configured" error so callers see a stable
// FailedPrecondition rather than a nil pointer crash, and the binary boots
// without FreeSWITCH being reachable.
type NoopPBX struct{}

func (NoopPBX) Name() string    { return "noop" }
func (NoopPBX) Enabled() bool   { return false }

var errPBXNotConfigured = errors.New("telephony: PBX controller not configured")

func (NoopPBX) Originate(context.Context, OriginateRequest) (string, error) {
	return "", errPBXNotConfigured
}
func (NoopPBX) Hangup(context.Context, string, string) error { return errPBXNotConfigured }
func (NoopPBX) Transfer(context.Context, string, string, bool) (string, error) {
	return "", errPBXNotConfigured
}
func (NoopPBX) CompleteTransfer(context.Context, string, string) error { return errPBXNotConfigured }
func (NoopPBX) Hold(context.Context, string) error                     { return errPBXNotConfigured }
func (NoopPBX) Resume(context.Context, string) error                   { return errPBXNotConfigured }
func (NoopPBX) ProvisionExtension(context.Context, SipExtension, string) error {
	return errPBXNotConfigured
}
func (NoopPBX) ProvisionTrunk(context.Context, SipTrunk) error { return errPBXNotConfigured }

func (NoopPBX) PlayAudio(context.Context, string, string) error              { return errPBXNotConfigured }
func (NoopPBX) SayText(context.Context, string, string, string) error        { return errPBXNotConfigured }
func (NoopPBX) PlayAndGetDigits(context.Context, string, PlayAndGetDigitsOpts) error {
	return errPBXNotConfigured
}
func (NoopPBX) StartRecording(context.Context, string, string, int) error { return errPBXNotConfigured }
