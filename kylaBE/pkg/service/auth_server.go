package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"log"
	"strings"
	"time"

	"kyla-be/pkg/pb"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthServer is the server API for AuthService service.
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	UserStore    *UserStore
	AuthStore    *AuthStore
	DBAuthStore  *DBAuthStore
	jwtManager   *JWTManager
	RbacStore    *RbacStore
	EmailService *utils.ResendService
	FirebaseAuth *FirebaseAuthService
	webAuthn     *webauthn.WebAuthn
}

// NewAuthServer creates a new instance of AuthServer.
func NewAuthServer(
	UserStore *UserStore,
	jwtManager *JWTManager,
	AuthStore *AuthStore,
	RbacStore *RbacStore,
	EmailService *utils.ResendService,
	FirebaseAuth *FirebaseAuthService,
	webAuthn *webauthn.WebAuthn,
	DBAuthStore *DBAuthStore,
) *AuthServer {
	return &AuthServer{
		UnimplementedAuthServiceServer: pb.UnimplementedAuthServiceServer{},
		UserStore:                      UserStore,
		jwtManager:                     jwtManager,
		AuthStore:                      AuthStore,
		RbacStore:                      RbacStore,
		EmailService:                   EmailService,
		FirebaseAuth:                   FirebaseAuth,
		webAuthn:                       webAuthn,
		DBAuthStore:                    DBAuthStore,
	}
}

// Login performs user authentication and generates an access token.
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Println("Login")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to login")
	}

	email := req.GetEmail()
	username := req.GetUsername()
	password := req.GetPassword()

	user := &User{}
	if email == "" {
		thisUser, err := s.UserStore.FindByUsername(username)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
		}
		user = thisUser
	} else if username == "" {
		thisUser, err := s.UserStore.FindByEmail(email)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
		}
		user = thisUser
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "email or username is required")
	}

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	if strings.EqualFold(user.Status, k.USER_STATUSES()["INACTIVE"]) {
		return nil, status.Errorf(codes.PermissionDenied, "user is not active")
	}
	if !utils.COMPARE_PASSWORD(user.HashedPassword, password) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid Credentials, Please check and try again")
	}

	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	ipAddress := GetIPAddress(ctx)

	// create a session
	session := &UserSession{
		ID:           uuid.New(),
		UserID:       user.ID,
		StartTime:    time.Now().Format(time.RFC3339),
		EndTime:      time.Now().Add(time.Hour * 24 * 7).Format(time.RFC3339), // 7 days session
		IpAddress:    ipAddress,
		IsValid:      true,
		LastLoggedIn: time.Now().Format(time.RFC3339),
	}
	sess, err := s.DBAuthStore.CreateSession(session)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create session: %v", err)
	}
	org, err := s.AuthStore.OrganisationStore.FindByID(&user.CurrentOrganisationID)
	if err != nil {
		org = &Organisation{}
	}

	res := &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToPbUser(user),
		UserSession:  SessionToPbUserSession(sess),
		Organisation: OrganisationToPbOrganisation(org),
	}
	return res, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	log.Println("Refresh Token")

	refreshToken := req.GetRefreshToken()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "metadata is not provided")
	}

	accessclaims, err := s.jwtManager.ValidateToken(md["authorization"][0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	claims, err := s.jwtManager.ValidateRefreshToken(accessclaims, refreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh token is invalid: %v", err)
	}

	userID := uuid.MustParse(claims.UserId)
	user, err := s.UserStore.FindByID(&userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	res := &pb.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
	return res, nil
}

// ChangePassword handles the ChangePassword RPC
func (s *AuthServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	log.Println("Change Password")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: Could not verify user to complete your request")
	}

	userID := contextData.UserID
	oldPassword := req.GetOldPassword()
	newPassword := req.GetNewPassword()

	// Assuming you have a function to get the user's current password from the database
	currentPassword, err := s.UserStore.ReadUserPasswordByID(userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get current password: %v", err)
	}

	// Check if the old password matches the one stored in the database
	if err := utils.COMPARE_PASSWORD(currentPassword, oldPassword); !err {
		return nil, status.Errorf(403, "invalid password, please try again")
	}

	// Call the new ChangePassword method in the UserStore
	user, err := s.UserStore.ChangePassword(&userID, newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to change password: %v", err)
	}

	// Send a confirmation email (you need to implement this function)
	if err := s.EmailService.SEND_PASSWORD_CHANGE_CONFIRMATION_EMAIL(user.FirstName, user.Email); err != nil {
		log.Printf("failed to send password change confirmation email: %v", err)
	}

	return &pb.ChangePasswordResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Password changed successfully",
		},
	}, nil
}

// ForgotPassword is the RPC to handle forgot password requests
func (s *AuthServer) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	// Logic to find the user by username, email, user ID, or old user ID
	user, err := s.UserStore.FindUserByForgotPasswordRequest(req)
	if err != nil {
		return nil, err
	}

	// Generate a new random password
	newPassword, err := utils.GENERATE_RANDOM_KEY(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new password: %v", err)
	}

	user.HashedPassword = utils.HASH_PASSWORD(newPassword)
	// Save the updated user
	if _, err := s.UserStore.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	// Send the new password to the user (through email)
	err = s.EmailService.SEND_NEW_PASSWORD_EMAIL(user.Username, user.Email, newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to send new password email: %v", err)
	}

	return &pb.ForgotPasswordResponse{
		Message: "New password sent to the user's email",
	}, nil
}

func (s *AuthServer) ReadAuthContext(_ context.Context, req *pb.ReadAuthContextRequest) (*pb.ReadAuthContextResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic: ", r)
		}
	}()

	log.Println("AuthContext Performed by: ", req.GetMethodName())
	authorization := req.GetAccessToken()
	method := req.GetMethodName()
	codeName := k.ROUTE_PERMISSIONS()[method]
	if codeName == "" {
		codeName = k.OPEN_ROUTES()[method]
		if codeName == "" {
			return nil, status.Errorf(400, "invalid method name")
		}
	}
	contextData, err := s.AuthStore.AuthInternalRequests(authorization, codeName)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	roles := []string{}
	for _, role := range contextData.Roles {
		roles = append(roles, role.String())
	}
	idNameMappings := []*pb.IdNameMapping{}
	for _, idNameMapping := range contextData.IdNameMappings {
		idNameMappings = append(idNameMappings, &pb.IdNameMapping{
			Id:   idNameMapping.ID,
			Name: idNameMapping.Name,
		})
	}
	if k.OPEN_ROUTES()[method] != "" {
		contextData.RequestAuth = k.NewConsts().TRUE
	}

	user, err := s.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		log.Println("Error fetching user: ", err)
		return nil, status.Errorf(codes.Internal, "error fetching user: %v", err)
	}

	return &pb.ReadAuthContextResponse{
		Metadata: &pb.RequestMetadata{
			Authorization:  contextData.Authorization,
			OrganisationId: contextData.OrganisationID.String(),
			UserId:         contextData.UserID.String(),
			BranchId:       contextData.BranchID.String(),
			RequestAuth:    contextData.RequestAuth,
			Roles:          roles,
			User:           UserToPbUser(user),
			AuthScope: &pb.AuthScope{
				UserId:         contextData.Scopes.User.String(),
				OrganisationId: contextData.Scopes.Organisation.String(),
				CurrentBranch:  contextData.Scopes.Branch.String(),
				Teams:          contextData.Scopes.Teams,
				Departments:    contextData.Scopes.Departments,
				Branches:       contextData.Scopes.Branches,
			},
			IdNameMappings: idNameMappings,
		},
	}, nil
}

// get permissions belonging to a user
func (s *AuthServer) ReadUserPermissions(ctx context.Context, request *pb.ReadUserPermissionsRequest) (*pb.ReadUserPermissionsResponse, error) {
	log.Println("Read User Permissions")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to get permissions resource")
	}

	user, err := s.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Error(500, "error while fetching user")
	}
	roleIDs := []uuid.UUID{}
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}
	roles, err := s.RbacStore.FindRolesByIDs(roleIDs)

	if err != nil {
		return nil, status.Error(500, "error while fetching permissions")
	}

	permissions := ReadPermissionsFromRoles(roles)

	return &pb.ReadUserPermissionsResponse{
		Permissions: permissions,
	}, nil
}

func (s *AuthServer) CheckUserPermission(ctx context.Context, req *pb.CheckUserPermissionRequest) (*pb.CheckUserPermissionResponse, error) {
	log.Println("Check User Permission")
	userId, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	user, err := s.UserStore.FindByID(&userId)
	if err != nil {
		return nil, status.Error(500, "error while fetching user")
	}

	roleIDs := []uuid.UUID{}
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}
	roles, err := s.RbacStore.FindRolesByIDs(roleIDs)
	if err != nil {
		return nil, status.Error(500, "error while fetching roles")
	}

	granted := false
	scopes := []*pb.Scope{}

	for _, role := range roles {
		for _, perm := range role.PermissionCodeNames {
			if strings.EqualFold(perm, req.Permission) {
				granted = true
				scopes = append(scopes, &pb.Scope{
					OwnerType: pb.OwnerType(pb.OwnerType_value[role.OwnerType.String()]),
					OwnerId:   role.OwnerID.String(),
				})
				break
			}
		}
	}
	// Check if user has the required permission
	return &pb.CheckUserPermissionResponse{
		Granted: granted,
		Scopes:  scopes,
	}, nil
}

// GoogleSignIn handles authentication via Google ID token
func (s *AuthServer) GoogleSignIn(ctx context.Context, req *pb.GoogleSignInRequest) (*pb.LoginResponse, error) {
	log.Println("Google Sign In")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to login")
	}

	idToken := req.GetIdToken()
	if idToken == "" {
		return nil, status.Error(codes.InvalidArgument, "ID token is required")
	}

	// Verify the ID token with Firebase
	firebaseToken, err := s.FirebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid ID token: %v", err)
	}

	// Find or create user based on Firebase UID
	user, err := s.UserStore.FindByFirebaseUID(firebaseToken.UID)
	if err != nil {
		// Generate a random password
		password, err := utils.GENERATE_RANDOM_KEY(16)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate password: %v", err)
		}

		// If user doesn't exist, create a new one
		user = &User{
			ID:             uuid.New(),
			Email:          firebaseToken.Claims["email"].(string),
			Username:       firebaseToken.Claims["email"].(string),
			FirstName:      firebaseToken.Claims["name"].(string),
			Status:         k.USER_STATUSES()["ACTIVE"],
			FirebaseUID:    firebaseToken.UID,
			HashedPassword: utils.HASH_PASSWORD(password),
		}

		user, err = s.UserStore.Create(user)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	}

	if strings.EqualFold(user.Status, k.USER_STATUSES()["INACTIVE"]) {
		return nil, status.Error(codes.PermissionDenied, "user is not active")
	}

	// Generate JWT tokens
	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToPbUser(user),
	}, nil
}

// MicrosoftSignIn handles authentication via Microsoft ID token
func (s *AuthServer) MicrosoftSignIn(ctx context.Context, req *pb.MicrosoftSignInRequest) (*pb.LoginResponse, error) {
	log.Println("Microsoft Sign In")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to login")
	}

	idToken := req.GetIdToken()
	if idToken == "" {
		return nil, status.Error(codes.InvalidArgument, "ID token is required")
	}

	// Handle Microsoft sign-in using the common social sign-in handler
	user, err := s.FirebaseAuth.HandleSocialSignIn(ctx, idToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid ID token: %v", err)
	}

	if user.Status != k.USER_STATUSES()["ACTIVE"] {
		return nil, status.Error(codes.PermissionDenied, "user is not active")
	}

	// Generate JWT tokens
	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToPbUser(user),
	}, nil
}

// FacebookSignIn handles authentication via Facebook ID token
func (s *AuthServer) FacebookSignIn(ctx context.Context, req *pb.FacebookSignInRequest) (*pb.LoginResponse, error) {
	log.Println("Facebook Sign In")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to login")
	}

	idToken := req.GetIdToken()
	if idToken == "" {
		return nil, status.Error(codes.InvalidArgument, "ID token is required")
	}

	// Handle Facebook sign-in using the common social sign-in handler
	user, err := s.FirebaseAuth.HandleSocialSignIn(ctx, idToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid ID token: %v", err)
	}

	if strings.EqualFold(user.Status, k.USER_STATUSES()["INACTIVE"]) {
		return nil, status.Error(codes.PermissionDenied, "user is not active")
	}

	// Generate JWT tokens
	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToPbUser(user),
	}, nil
}

// MFASetup handles the setup of MFA for a user
func (s *AuthServer) MFASetup(ctx context.Context, req *pb.MFASetupRequest) (*pb.MFASetupResponse, error) {
	// Check if user exists
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	user, err := s.UserStore.FindByID(&userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Generate TOTP secret
	secret, err := GenerateMFASecret()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate MFA secret: %v", err)
	}

	// Generate QR code
	qrCode, err := GenerateQRCode(secret, user.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate QR code: %v", err)
	}

	// Generate recovery codes
	recoveryCodes, err := GenerateRecoveryCodes(8)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate recovery codes: %v", err)
	}

	// Update user with MFA details
	user.MFAEnabled = true
	user.MFASecret = secret
	user.MFARecoveryCodes = recoveryCodes

	_, err = s.UserStore.Update(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.MFASetupResponse{
		Secret:        secret,
		QrCode:        qrCode,
		RecoveryCodes: recoveryCodes,
	}, nil
}

// MFAVerify verifies a TOTP code for a user
func (s *AuthServer) MFAVerify(ctx context.Context, req *pb.MFAVerifyRequest) (*pb.MFAVerifyResponse, error) {
	// Check if user exists
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	user, err := s.UserStore.FindByID(&userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Verify TOTP code
	if !VerifyTOTP(user.MFASecret, req.Code) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid TOTP code")
	}

	// Update last MFA login time
	user.LastMFALogin = time.Now()
	_, err = s.UserStore.Update(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.MFAVerifyResponse{
		Success: true,
		Message: "MFA verification successful",
	}, nil
}

// PasskeyRegistration handles passkey registration for a user
func (s *AuthServer) PasskeyRegistration(ctx context.Context, req *pb.PasskeyRegistrationRequest) (*pb.PasskeyRegistrationResponse, error) {
	// Check if user exists
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	user, err := s.UserStore.FindByID(&userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Create WebAuthn user
	webAuthnUser := &UserWebAuthn{
		ID:          []byte(user.ID.String()),
		Name:        user.Email,
		DisplayName: user.FirstName,
	}

	// Begin registration
	options, session, err := s.webAuthn.BeginRegistration(webAuthnUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin passkey registration: %v", err)
	}

	// Store session data
	if err := s.storePasskeySession(user.ID.String(), session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store passkey session: %v", err)
	}

	// Convert options to JSON for the client
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal options: %v", err)
	}

	return &pb.PasskeyRegistrationResponse{
		CredentialId: string(optionsJSON), // Send the entire options object as JSON
		PublicKey:    "",                  // The client will handle the public key from the options
	}, nil
}

// storePasskeySession stores the WebAuthn session data
func (s *AuthServer) storePasskeySession(userID string, session *webauthn.SessionData) error {
	// Store session data in Redis with a 5-minute expiration
	key := fmt.Sprintf("webauthn:session:%s", userID)
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %v", err)
	}

	// Store in Redis with 5-minute expiration
	err = s.AuthStore.RedisClient.Set(context.Background(), key, sessionJSON, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to store session data: %v", err)
	}

	return nil
}

// getPasskeySession retrieves the WebAuthn session data
func (s *AuthServer) getPasskeySession(userID string) (*webauthn.SessionData, error) {
	key := fmt.Sprintf("webauthn:session:%s", userID)

	// Get session data from Redis
	sessionJSON, err := s.AuthStore.RedisClient.Get(context.Background(), key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %v", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %v", err)
	}

	return &session, nil
}

// LoginWithMFA handles login with MFA verification
func (s *AuthServer) LoginWithMFA(ctx context.Context, req *pb.LoginWithMFARequest) (*pb.LoginResponse, error) {
	// Check if user exists
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	user, err := s.UserStore.FindByID(&userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Verify that MFA is enabled for the user
	if !user.MFAEnabled {
		return nil, status.Errorf(codes.FailedPrecondition, "MFA is not enabled for this user")
	}

	// Verify TOTP code
	if !VerifyTOTP(user.MFASecret, req.Code) {
		// Check if it's a recovery code
		isRecoveryCode := false
		for i, code := range user.MFARecoveryCodes {
			if code == req.Code {
				isRecoveryCode = true
				// Remove used recovery code
				user.MFARecoveryCodes = append(user.MFARecoveryCodes[:i], user.MFARecoveryCodes[i+1:]...)
				break
			}
		}
		if !isRecoveryCode {
			return nil, status.Errorf(codes.InvalidArgument, "invalid MFA code")
		}
	}

	// Update last MFA login time
	user.LastMFALogin = time.Now()
	_, err = s.UserStore.Update(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToPbUser(user),
	}, nil
}

// LoginWithPasskey handles login with passkey authentication
func (s *AuthServer) LoginWithPasskey(ctx context.Context, req *pb.LoginWithPasskeyRequest) (*pb.LoginResponse, error) {
	// Check if user exists
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	user, err := s.UserStore.FindByID(&userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Create WebAuthn user
	webAuthnUser := &UserWebAuthn{
		ID:          []byte(user.ID.String()),
		Name:        user.Email,
		DisplayName: user.FirstName,
	}

	// Get stored session data
	session, err := s.getPasskeySession(user.ID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get passkey session: %v", err)
	}

	// Parse the credential response from the client
	credentialBytes := []byte(req.GetChallenge())
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(credentialBytes))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid credential response: %v", err)
	}

	// Find the credential
	var credential *webauthn.Credential
	for _, passkey := range user.Passkeys {
		if string(passkey.CredentialID) == req.GetCredentialId() {
			credential = &webauthn.Credential{
				ID:        passkey.CredentialID,
				PublicKey: passkey.PublicKey,
				Authenticator: webauthn.Authenticator{
					AAGUID:       passkey.Authenticator.AAGUID,
					SignCount:    passkey.Authenticator.SignCount,
					CloneWarning: passkey.Authenticator.CloneWarning,
				},
			}
			break
		}
	}

	if credential == nil {
		return nil, status.Errorf(codes.NotFound, "passkey not found")
	}

	// Verify the assertion
	_, err = s.webAuthn.ValidateLogin(webAuthnUser, *session, parsedResponse)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "passkey authentication failed: %v", err)
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToPbUser(user),
	}, nil
}
