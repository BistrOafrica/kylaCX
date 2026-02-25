package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the server implementation for the Auth service.
type UserServer struct {
	pb.UnimplementedUserServiceServer
	UserStore        *UserStore
	AuthStore        *AuthStore
	AgentStatusStore *StatusStore
	EmailService     *utils.ResendService
}

func NewUserServer(
	UserStore *UserStore,
	AuthStore *AuthStore,
	AgentAgentStatusStore *StatusStore,
	EmailService *utils.ResendService,
) *UserServer {
	return &UserServer{
		UserStore:        UserStore,
		AuthStore:        AuthStore,
		AgentStatusStore: AgentAgentStatusStore,
		EmailService:     EmailService,
	}
}

// CreateUser creates a new user in the database.
func (o *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Println("Create User")

	scope := PbScopeToOpScope(req.GetScope())
	contextData, err := o.AuthStore.GetServiceAuthMetadata(ctx)
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

	// Check if the email and username are unique
	if o.UserStore.CheckUserExist(user.Email, user.Username, user.Phone) {
		return nil, status.Error(500, "email or username already exists")
	}

	if len(user.Roles) == 0 {
		agentRole, err := o.AuthStore.RbacStore.FindByName("Agent", scope)
		if err != nil {
			log.Printf("Error while fetching agent role: %v", err)
		}
		user.Roles = append(user.Roles, *agentRole)
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

	user.ADD_BASIC_ROLE(o.AuthStore.RbacStore)
	if err := user.CREATE_AGENT_STATUS(o.AgentStatusStore); err != nil {
		return nil, status.Error(500, "error while creating agent status")
	}
	newUser, err := o.UserStore.Save(user, true)
	if err != nil {
		return nil, status.Error(500, "error while saving user")
	}

	if emailErr := o.EmailService.SEND_WELCOME_EMAIL(newUser.Email, newUser.Username, newUser.FirstName, password); emailErr != nil {
		log.Printf("Error sending welcome email: %v", err)
	}

	return &pb.CreateUserResponse{
		User: UserToPbUser(newUser),
	}, nil
}

func (o *UserServer) ReadUser(ctx context.Context, request *pb.ReadUserRequest) (*pb.ReadUserResponse, error) {
	log.Println("Read User")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(request.GetScope())
	go o.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			log.Println("Request Auth: ", contextChanData.RequestAuth)
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
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

func (o *UserServer) ReadUsers(ctx context.Context, req *pb.ReadUsersRequest) (*pb.ReadUsersResponse, error) {
	log.Println("Read all users")
	// contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go o.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		// if !CheckOpScope(contextChanData, scope) {
		// 	return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		// }
		// contextData = contextChanData
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

func (o *UserServer) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log.Println("Update user")

	contextData, err := o.AuthStore.GetServiceAuthMetadata(ctx)
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

// DeleteUser deletes an existing user from the database.
func (o *UserServer) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Println("Delete user")

	contextData, err := o.AuthStore.GetServiceAuthMetadata(ctx)
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

func (o *UserServer) AddRoleToUser(ctx context.Context, request *pb.RoleToUserRequest) (*pb.RoleToUserResponse, error) {
	log.Println("Add role to user")

	// contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go o.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		// contextData = contextChanData
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

func (o *UserServer) RemoveRoleFromUser(ctx context.Context, request *pb.RoleToUserRequest) (*pb.RoleToUserResponse, error) {
	log.Println("Remove role from user")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go o.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

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

func (o *UserServer) AddUsersToRole(ctx context.Context, request *pb.UsersToRoleRequest) (*pb.UsersToRoleResponse, error) {
	log.Println("Add users to role")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go o.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
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

func (o *UserServer) DeactivateUser(ctx context.Context, request *pb.DeactivateUserRequest) (*pb.DeactivateUserResponse, error) {
	log.Println("Deactivate user")

	contextData, err := o.AuthStore.GetServiceAuthMetadata(ctx)
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

func (o *UserServer) ActivateUser(ctx context.Context, request *pb.ActivateUserRequest) (*pb.ActivateUserResponse, error) {
	log.Println("Activate user")

	contextData, err := o.AuthStore.GetServiceAuthMetadata(ctx)
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

func (o *UserServer) SearchUsers(ctx context.Context, request *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
	// contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go o.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to search users resource")
		}
		// contextData = contextChanData
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

	query := request.GetQuery()

	users, count, err := o.UserStore.Search(query, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("error while searching users: %v", err)
	}

	return &pb.SearchUsersResponse{
		Users: UsersToPbUsers(users),
		Count: count,
	}, nil
}

// Read users by permission codename
func (r *RbacServer) ReadUsersByPermission(_ context.Context, request *pb.ReadUsersByPermissionRequest) (*pb.ReadUsersByPermissionResponse, error) {
	log.Println("Read Users By Permission")
	scope := PbScopeToOpScope(request.GetScope())
	roleChan := make(chan []string)
	errChan := make(chan error)
	users := []*pb.User{}
	go r.getPermissionRoles(request.GetCodeName(), scope, roleChan, errChan)

	select {
	case roleIds := <-roleChan:
		usersChan := make(chan []*pb.User)
		usersErrChan := make(chan error)
		go r.getRoleUsers(roleIds, usersChan, usersErrChan)
		select {
		case usersData := <-usersChan:
			users = append(users, usersData...)
		case err := <-usersErrChan:
			return nil, status.Errorf(codes.Internal, "Failed to get users by role %v", err)
		}
	case errChan := <-errChan:
		return nil, status.Errorf(codes.Internal, "Failed to get roles by permission %v", errChan)
	}

	return &pb.ReadUsersByPermissionResponse{
		Users: users,
	}, nil
}

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

func (o *UserServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	// Validate required fields
	if req.GetEmail() == "" || req.GetFirstName() == "" || req.GetLastName() == "" || req.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, first name, last name, and username are required")
	}

	// Validate terms acceptance
	if !req.GetTerms() {
		return nil, status.Errorf(codes.InvalidArgument, "terms must be accepted")
	}

	// Create new user
	// Validate email format
	if !utils.ValidateEmailFormat(req.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	// Validate phone format if provided
	if req.GetPhone() != "" && !utils.ValidatePhoneFormat(req.GetPhone()) {
		return nil, status.Error(codes.InvalidArgument, "invalid phone format")
	}

	// Check if email/username already exists
	if o.UserStore.CheckUserExist(req.GetEmail(), req.GetUsername(), req.GetPhone()) {
		return nil, status.Error(codes.AlreadyExists, "email or username already exists")
	}

	password, err := utils.GENERATE_RANDOM_KEY(10)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate password")
	}

	// Create user object
	user := &User{
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

	// Add basic role
	user.CREATE_NEW_USER_ROLE(o.AuthStore.RbacStore)

	// Create agent status
	if err := user.CREATE_AGENT_STATUS(o.AgentStatusStore); err != nil {
		return nil, status.Error(codes.Internal, "error while creating agent status")
	}

	// Save user
	newUser, err := o.UserStore.Save(user, true)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Send welcome email
	if emailErr := o.EmailService.SEND_WELCOME_EMAIL(newUser.Email, newUser.Username, newUser.FirstName, password); emailErr != nil {
		log.Printf("Error sending welcome email: %v", emailErr)
	}

	return &pb.SignUpResponse{
		UserId: newUser.ID.String(),
		Status: &pb.Status{
			Code:    200,
			Message: "Your Account has been created successfully. Check your email for verification.",
		},
	}, nil
}

// activate user account
func (o *UserServer) ActivateUserAccount(ctx context.Context, req *pb.ActivateUserAccountRequest) (*pb.ActivateUserAccountResponse, error) {
	// Validate required fields
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}
	// Activate user account
	user, err := o.UserStore.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Activate user by updating status in the store
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
