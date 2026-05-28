package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyla-be/internal/agentops"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway abstracts auth-metadata resolution needed by UserServer.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	GetUserRequestMetadata(ctx context.Context, dataChan chan *authctx.RequestMetadata, errChan chan error)
}

// UserServer implements the User gRPC service.
type UserServer struct {
	pb.UnimplementedUserServiceServer
	UserStore        *UserStore
	AuthGateway      AuthGateway
	RbacStore        *rbac.RbacStore
	AgentStatusStore *agentops.StatusStore
	EmailService     *utils.ResendService
}

// NewUserServer constructs a new UserServer.
func NewUserServer(
	userStore *UserStore,
	authGateway AuthGateway,
	rbacStore *rbac.RbacStore,
	agentStatusStore *agentops.StatusStore,
	emailService *utils.ResendService,
) *UserServer {
	return &UserServer{
		UserStore:        userStore,
		AuthGateway:      authGateway,
		RbacStore:        rbacStore,
		AgentStatusStore: agentStatusStore,
		EmailService:     emailService,
	}
}

// CreateUser creates a new user in the database.
func (o *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Println("Create User")

	scope := authctx.PbScopeToOpScope(req.GetScope())
	contextData, err := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to create user resource")
	}
	user := PbUserToUser(req.GetUser())

	if user.Username == "" {
		return nil, status.Error(500, "username is required")
	}
	if !utils.ValidateEmailFormat(user.Email) {
		return nil, status.Error(500, "invalid email format")
	}
	if !utils.ValidatePhoneFormat(user.Phone) {
		return nil, status.Error(500, "invalid phone format")
	}
	if o.UserStore.CheckUserExist(user.Email, user.Username, user.Phone) {
		return nil, status.Error(500, "email or username already exists")
	}

	if len(user.Roles) == 0 {
		agentRole, err := o.RbacStore.FindByName("Agent", scope)
		if err != nil {
			log.Printf("Error while fetching agent role: %v", err)
		} else {
			user.Roles = append(user.Roles, *agentRole)
		}
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.CreatedBy = contextData.UserID.String()
	user.Status = k.USER_STATUSES()["NEW"]
	password, err := utils.GENERATE_RANDOM_KEY(10)
	if err != nil {
		return nil, status.Error(500, "error while generating password")
	}
	user.HashedPassword = utils.HASH_PASSWORD(password)

	user.ADD_BASIC_ROLE(o.RbacStore)
	if err := user.CREATE_AGENT_STATUS(o.AgentStatusStore); err != nil {
		return nil, status.Error(500, "error while creating agent status")
	}
	newUser, err := o.UserStore.Save(user, true)
	if err != nil {
		return nil, status.Error(500, "error while saving user")
	}

	if emailErr := o.EmailService.SEND_WELCOME_EMAIL(newUser.Email, newUser.Username, newUser.FirstName, password); emailErr != nil {
		log.Printf("Error sending welcome email: %v", emailErr)
	}

	return &pb.CreateUserResponse{
		User: UserToPbUser(newUser),
	}, nil
}

// ReadUser retrieves a single user by ID.
func (o *UserServer) ReadUser(ctx context.Context, request *pb.ReadUserRequest) (*pb.ReadUserResponse, error) {
	log.Println("Read User")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	scope := authctx.PbScopeToOpScope(request.GetScope())
	go o.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			log.Println("Request Auth: ", contextChanData.RequestAuth)
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		if !authctx.CheckOpScope(contextChanData, scope) {
			log.Println("Scope: ", scope)
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	userID, err := uuid.Parse(request.GetUser())
	if err != nil {
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}
	user, err := o.UserStore.FindByID(&userID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching user: %v", err)
	}

	return &pb.ReadUserResponse{
		User: UserToPbUser(user),
	}, nil
}

// ReadUsers retrieves all users matching the request scope.
func (o *UserServer) ReadUsers(ctx context.Context, req *pb.ReadUsersRequest) (*pb.ReadUsersResponse, error) {
	log.Println("Read all users")
	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	scope := authctx.PbScopeToOpScope(req.GetScope())
	go o.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		_ = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	users, err := o.UserStore.FindByScope(scope)
	if err != nil {
		return nil, status.Error(500, "error while fetching users")
	}

	return &pb.ReadUsersResponse{
		Users: UsersToPbUsers(users),
		Status: &pb.Status{
			Code:    200,
			Message: "Success",
		},
	}, nil
}

// UpdateUser updates an existing user record.
func (o *UserServer) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log.Println("Update user")

	contextData, err := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to update user resource")
	}

	user := PbUserToUser(request.GetUser())
	updatedUser, err := o.UserStore.Update(user)
	if err != nil {
		return nil, fmt.Errorf("error while updating user: %v", err)
	}

	return &pb.UpdateUserResponse{
		User: UserToPbUser(updatedUser),
	}, nil
}

// DeleteUser deletes an existing user.
func (o *UserServer) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Println("Delete user")

	contextData, err := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to delete user resource")
	}

	if err := o.UserStore.Delete(request.GetId()); err != nil {
		return nil, fmt.Errorf("error while deleting user: %v", err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}

// ReadAllUsersWithoutToken retrieves all users without requiring an auth token.
func (o *UserServer) ReadAllUsersWithoutToken(_ context.Context, _ *pb.ReadAllUsersWithoutTokenRequest) (*pb.ReadAllUsersWithoutTokenResponse, error) {
	log.Println("Read all users")

	users, err := o.UserStore.RetrieveAllUsers()
	if err != nil {
		return nil, status.Error(500, "error while fetching users")
	}

	return &pb.ReadAllUsersWithoutTokenResponse{
		Users: UsersToPbUsers(users),
		Status: &pb.Status{
			Code:    200,
			Message: "Success",
		},
	}, nil
}

// AddRoleToUser adds a role to a user.
func (o *UserServer) AddRoleToUser(ctx context.Context, request *pb.RoleToUserRequest) (*pb.RoleToUserResponse, error) {
	log.Println("Add role to user")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go o.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	user, err := o.UserStore.AddRoleToUser(request.GetUserId(), request.GetRoleId())
	if err != nil {
		return nil, status.Error(500, "error while adding role to user")
	}

	return &pb.RoleToUserResponse{
		User: UserToPbUser(user),
	}, nil
}

// RemoveRoleFromUser removes a role from a user.
func (o *UserServer) RemoveRoleFromUser(ctx context.Context, request *pb.RoleToUserRequest) (*pb.RoleToUserResponse, error) {
	log.Println("Remove role from user")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go o.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	user, err := o.UserStore.RemoveRoleFromUser(request.GetUserId(), request.GetRoleId())
	if err != nil {
		return nil, status.Error(500, "error while removing role from user")
	}

	return &pb.RoleToUserResponse{
		User: UserToPbUser(user),
	}, nil
}

// AddUsersToRole assigns a role to multiple users.
func (o *UserServer) AddUsersToRole(ctx context.Context, request *pb.UsersToRoleRequest) (*pb.UsersToRoleResponse, error) {
	log.Println("Add users to role")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go o.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		_ = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	if err := o.UserStore.AddUsersToRole(request.GetRoleId(), request.GetUserIds()); err != nil {
		return nil, status.Error(500, "error while adding users to role")
	}

	return &pb.UsersToRoleResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Successfully added users to role",
		},
	}, nil
}

// DeactivateUser sets a user's status to INACTIVE.
func (o *UserServer) DeactivateUser(ctx context.Context, request *pb.DeactivateUserRequest) (*pb.DeactivateUserResponse, error) {
	log.Println("Deactivate user")

	contextData, err := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to deactivate user resource")
	}

	userID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}
	if err := o.UserStore.DeactivateUser(&userID); err != nil {
		return nil, fmt.Errorf("error while deactivating user: %v", err)
	}

	return &pb.DeactivateUserResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Successfully deactivated user",
		},
	}, nil
}

// ActivateUser sets a user's status to ACTIVE.
func (o *UserServer) ActivateUser(ctx context.Context, request *pb.ActivateUserRequest) (*pb.ActivateUserResponse, error) {
	log.Println("Activate user")

	contextData, err := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to activate user resource")
	}

	userID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}
	if err := o.UserStore.ActivateUser(&userID); err != nil {
		return nil, fmt.Errorf("error while activating user: %v", err)
	}

	return &pb.ActivateUserResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Successfully activated user",
		},
	}, nil
}

// SearchUsers performs a fuzzy search over users.
func (o *UserServer) SearchUsers(ctx context.Context, request *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go o.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to search users resource")
		}
		_ = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	page := request.GetPage()
	if page == 0 {
		page = 1
	}
	perPage := request.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	users, count, err := o.UserStore.Search(request.GetQuery(), page, perPage)
	if err != nil {
		return nil, fmt.Errorf("error while searching users: %v", err)
	}

	return &pb.SearchUsersResponse{
		Users: UsersToPbUsers(users),
		Count: count,
	}, nil
}

// ReadByIdsFromService retrieves users by a list of IDs from an internal service call.
func (o *UserServer) ReadByIdsFromService(ctx context.Context, request *pb.ReadUsersByIdsFromServiceRequest) (*pb.ReadUsersByIdsFromServiceResponse, error) {
	log.Println("Read Users from Internal Service")

	users, err := o.UserStore.FindByIdsFromService(request.GetUserIds())
	if err != nil {
		return nil, err
	}

	return &pb.ReadUsersByIdsFromServiceResponse{
		Users: UsersToPbUsers(users),
	}, nil
}

// SignUp creates a self-registered user account.
func (o *UserServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	if req.GetEmail() == "" || req.GetFirstName() == "" || req.GetLastName() == "" || req.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, first name, last name, and username are required")
	}
	if !req.GetTerms() {
		return nil, status.Errorf(codes.InvalidArgument, "terms must be accepted")
	}
	if !utils.ValidateEmailFormat(req.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}
	if req.GetPhone() != "" && !utils.ValidatePhoneFormat(req.GetPhone()) {
		return nil, status.Error(codes.InvalidArgument, "invalid phone format")
	}
	if o.UserStore.CheckUserExist(req.GetEmail(), req.GetUsername(), req.GetPhone()) {
		return nil, status.Error(codes.AlreadyExists, "email or username already exists")
	}

	password, err := utils.GENERATE_RANDOM_KEY(10)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate password")
	}

	newUser := &User{
		ID:             uuid.New(),
		Email:          req.GetEmail(),
		FirstName:      req.GetFirstName(),
		LastName:       req.GetLastName(),
		Username:       req.GetUsername(),
		Phone:          req.GetPhone(),
		Status:         k.USER_STATUSES()["NEW"],
		ReferralCode:   req.GetReferralCode(),
		HashedPassword: utils.HASH_PASSWORD(password),
	}

	newUser.CREATE_NEW_USER_ROLE(o.RbacStore)

	if err := newUser.CREATE_AGENT_STATUS(o.AgentStatusStore); err != nil {
		return nil, status.Error(codes.Internal, "error while creating agent status")
	}

	savedUser, err := o.UserStore.Save(newUser, true)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	if emailErr := o.EmailService.SEND_WELCOME_EMAIL(savedUser.Email, savedUser.Username, savedUser.FirstName, password); emailErr != nil {
		log.Printf("Error sending welcome email: %v", emailErr)
	}

	return &pb.SignUpResponse{
		UserId: savedUser.ID.String(),
		Status: &pb.Status{
			Code:    200,
			Message: "Your Account has been created successfully. Check your email for verification.",
		},
	}, nil
}

// ActivateUserAccount activates a user account with a new password.
func (o *UserServer) ActivateUserAccount(ctx context.Context, req *pb.ActivateUserAccountRequest) (*pb.ActivateUserAccountResponse, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	user, err := o.UserStore.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if err := o.UserStore.ActivateUser(&user.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to activate user account")
	}

	return &pb.ActivateUserAccountResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "User account activated successfully",
		},
	}, nil
}
