package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"kyla-be/config"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"

	"google.golang.org/api/option"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
)

type FirebaseAuthService struct {
	app       *firebase.App
	client    *auth.Client
	userStore *UserStore
}

func NewFirebaseAuthService(config *config.FirebaseConfig, userStore *UserStore) (*FirebaseAuthService, error) {
	credsJSON, er := base64.StdEncoding.DecodeString(config.FbCredentials)
	if er != nil {
		return nil, fmt.Errorf("failed to decode Firebase credentials: %w", er)
	}
	opt := option.WithCredentialsJSON(credsJSON)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting auth client: %v", err)
	}

	return &FirebaseAuthService{
		app:       app,
		client:    client,
		userStore: userStore,
	}, nil
}

// VerifyIDToken verifies the Firebase ID token and returns the token claims
func (s *FirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := s.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %v", err)
	}
	return token, nil
}

// HandleSocialSignIn handles the common logic for all social sign-in providers
func (s *FirebaseAuthService) HandleSocialSignIn(ctx context.Context, idToken string) (*User, error) {
	token, err := s.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %v", err)
	}

	// Check if user exists by Firebase UID
	user, err := s.userStore.FindByFirebaseUID(token.UID)
	if err != nil {
		// If user doesn't exist, create new user
		if err.Error() == "failed to find user by Firebase UID: record not found" {
			// Generate a random password for the user
			password, err := utils.GENERATE_RANDOM_KEY(16)
			if err != nil {
				return nil, fmt.Errorf("error generating password: %v", err)
			}

			// Get user info from token claims
			email, _ := token.Claims["email"].(string)
			name, _ := token.Claims["name"].(string)
			// picture, _ := token.Claims["picture"].(string)

			// Create new user
			newUser := &User{
				ID:             uuid.New(),
				Email:          email,
				Username:       email, // Use email as username for now
				FirstName:      name,
				HashedPassword: utils.HASH_PASSWORD(password),
				Status:         k.USER_STATUSES()["ACTIVE"],
				FirebaseUID:    token.UID,
			}

			user, err = s.userStore.Create(newUser)
			if err != nil {
				return nil, fmt.Errorf("error creating user: %v", err)
			}

			return user, nil
		}
		return nil, fmt.Errorf("error finding user: %v", err)
	}

	return user, nil
}
