package telephony

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// JWTTokenIssuer implements TokenIssuer using HS256-signed JWTs. The secret
// is the same JWT_SECRET_KEY the rest of the auth stack uses — keeping the
// signing key centralised means one rotation invalidates everything.
//
// The issued token is bound to (org_id, user_id, extension) so a stolen token
// can't be used to register a different extension. FreeSWITCH validates the
// JWT signature via a mod_jwt-style script (or out-of-band auth handler) at
// registration time — see deploy/freeswitch/README.md.
type JWTTokenIssuer struct {
	secret []byte
	issuer string
}

// NewJWTTokenIssuer constructs an issuer. issuer goes into the JWT iss claim
// — should be something like "kyla-be" so consumers can distinguish our
// tokens from other systems sharing the secret.
func NewJWTTokenIssuer(secret, issuer string) *JWTTokenIssuer {
	if issuer == "" {
		issuer = "kyla-be"
	}
	return &JWTTokenIssuer{secret: []byte(secret), issuer: issuer}
}

// IssueSoftphoneToken returns a signed JWT for the given subject. Claims
// include the SIP extension so a single-purpose validator at the PBX edge
// can permit registration without needing to call back into our backend.
func (i *JWTTokenIssuer) IssueSoftphoneToken(orgID, userID, extension string, ttl time.Duration) (string, error) {
	if len(i.secret) == 0 {
		return "", errors.New("jwt issuer: secret not configured")
	}
	if userID == "" || extension == "" {
		return "", errors.New("jwt issuer: user_id and extension are required")
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":       i.issuer,
		"sub":       userID,
		"org":       orgID,
		"extension": extension,
		"scope":     "softphone",
		"iat":       now.Unix(),
		"exp":       now.Add(ttl).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(i.secret)
}
