package service

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"time"

	"github.com/duo-labs/webauthn/webauthn"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GenerateMFASecret generates a new TOTP secret
func GenerateMFASecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %v", err)
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

// GenerateQRCode generates a QR code for the authenticator app
func GenerateQRCode(secret, email string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "DialAfrika",
		AccountName: email,
		Secret:      []byte(secret),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	return key.URL(), nil
}

// GenerateRecoveryCodes generates a set of recovery codes
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code := make([]byte, 8)
		for j := 0; j < 8; j++ {
			num, err := rand.Int(rand.Reader, big.NewInt(36))
			if err != nil {
				return nil, fmt.Errorf("failed to generate recovery code: %v", err)
			}
			if num.Int64() < 10 {
				code[j] = byte('0' + num.Int64())
			} else {
				code[j] = byte('A' + (num.Int64() - 10))
			}
		}
		codes[i] = string(code)
	}
	return codes, nil
}

// VerifyTOTP verifies a TOTP code
func VerifyTOTP(secret, code string) bool {
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false
	}
	return valid
}

// WebAuthnConfig holds the WebAuthn configuration
type WebAuthnConfig struct {
	RPID          string
	RPOrigin      string
	RPDisplayName string
}

// NewWebAuthn creates a new WebAuthn instance
func NewWebAuthn(config *WebAuthnConfig) (*webauthn.WebAuthn, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigin:      config.RPOrigin,
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %v", err)
	}

	return webAuthn, nil
}

// UserWebAuthn represents a user for WebAuthn operations
type UserWebAuthn struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

// WebAuthnID returns the user's ID
func (u *UserWebAuthn) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName returns the user's name
func (u *UserWebAuthn) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName returns the user's display name
func (u *UserWebAuthn) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon returns the user's icon URL
func (u *UserWebAuthn) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns the user's credentials
func (u *UserWebAuthn) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
