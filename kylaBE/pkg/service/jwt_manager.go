package service

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// JWTManager is the interface to manage JWT tokens
type JWTManager struct {
	secretKey            string
	tokenDuration        time.Duration
	refreshtokenDuration time.Duration
	RbacStore            *RbacStore
}

// UserClaims is a custom JWT claims structure that contains user information
type UserClaims struct {
	jwt.StandardClaims
	UserId         uuid.UUID `json:"user_id"`
	Roles          []string  `json:"roles"`
	OrganisationID uuid.UUID `json:"organisation_id"`
	BranchID       uuid.UUID `json:"branch_id"`
}

type SessionTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenClaims struct {
	jwt.StandardClaims
	UserId string `json:"user_id"`
}

// NewJWTManager creates a new JWTManager
func NewJWTManager(secretKey string, tokenDuration time.Duration, refreshTokenDuration time.Duration, RbacStore *RbacStore) *JWTManager {
	return &JWTManager{
		secretKey:            secretKey,
		tokenDuration:        tokenDuration,
		refreshtokenDuration: refreshTokenDuration,
		RbacStore:            RbacStore,
	}
}

// GenerateToken generates a new JWT token for the given user
func (manager *JWTManager) GenerateToken(user *User) (*SessionTokens, error) {
	userRoleIDs := []string{}
	for _, role := range user.Roles {
		userRoleIDs = append(userRoleIDs, role.ID.String())
	}
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		UserId:         user.ID,
		Roles:          userRoleIDs,
		BranchID:       user.CurrentBranchID,
		OrganisationID: user.CurrentOrganisationID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(manager.secretKey))
	if err != nil {
		return &SessionTokens{}, err
	}
	refreshTokenClaims := RefreshTokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.refreshtokenDuration).Unix(),
		},
		UserId: user.ID.String(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(manager.secretKey))
	if err != nil {
		return &SessionTokens{}, err
	}

	return &SessionTokens{
		AccessToken:  signedToken,
		RefreshToken: signedRefreshToken,
	}, nil
}

// ValidateToken validates the given JWT token
func (manager *JWTManager) ValidateToken(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(manager.secretKey), nil
	})
	if err != nil {
		log.Printf("Error while validating token: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (manager *JWTManager) ValidateRefreshToken(accessClaims *UserClaims, refreshToken string) (*RefreshTokenClaims, error) {
	if accessClaims == nil {
		return nil, fmt.Errorf("accessClaims is nil")
	}

	token, err := jwt.ParseWithClaims(refreshToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(manager.secretKey), nil
	})
	if err != nil {
		log.Printf("Error while validating token: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.UserId != accessClaims.UserId.String() {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
