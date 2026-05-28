package invitation

import (
	"context"
	"fmt"
	"kyla-be/internal/agentops"
	"kyla-be/internal/authctx"
	"kyla-be/internal/branch"
	"kyla-be/internal/department"
	"kyla-be/internal/organisation"
	"kyla-be/internal/rbac"
	"kyla-be/internal/team"
	"kyla-be/internal/user"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/templates"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the interface for retrieving service-level auth metadata.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// InvitationServer implements the InvitationService interface
type InvitationServer struct {
	pb.UnimplementedInvitationServiceServer
	repo              *InvitationStore
	emailService      *utils.ResendService
	organisationStore *organisation.OrganisationStore
	branchStore       *branch.BranchStore
	departmentStore   *department.DepartmentStore
	teamStore         *team.TeamStore
	rbacStore         *rbac.RbacStore
	userStore         *user.UserStore
	authStore         AuthGateway
	agentStatusStore  *agentops.StatusStore
	baseURL           string
}

// NewInvitationServer creates a new service instance
func NewInvitationServer(
	repo *InvitationStore,
	emailService *utils.ResendService,
	orgService *organisation.OrganisationStore,
	branchService *branch.BranchStore,
	departmentService *department.DepartmentStore,
	teamService *team.TeamStore,
	rbacService *rbac.RbacStore,
	userService *user.UserStore,
	authService AuthGateway,
	agentStatusStore *agentops.StatusStore,
	baseURL string,
) *InvitationServer {
	return &InvitationServer{
		repo:              repo,
		emailService:      emailService,
		organisationStore: orgService,
		branchStore:       branchService,
		departmentStore:   departmentService,
		teamStore:         teamService,
		rbacStore:         rbacService,
		userStore:         userService,
		authStore:         authService,
		agentStatusStore:  agentStatusStore,
		baseURL:           baseURL,
	}
}

// MARK: Create Invitation
// CreateInvitation creates a new invitation and sends an email
func (s *InvitationServer) CreateInvitation(ctx context.Context, req *pb.CreateInviteRequest) (*pb.Invitation, error) {
	log.Println("Create Invitation")
	// Check service-level authentication
	contextData, err := s.authStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(codes.PermissionDenied, "forbidden: you do not have access to create invitation resource")
	}

	// Validate organization exists
	org, err := s.organisationStore.FindByID(&contextData.OrganisationID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}

	if len(req.RoleIds) < 1 {
		roles, err := s.rbacStore.FindDefaultRole(org.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to find default roles: %v", err)
		}
		req.RoleIds = []string{roles.ID.String()}
	}

	// Set invitation type based on userId presence
	if req.UserId != "" {
		// User ID provided - set type to existing user
		req.Type = pb.InvitationType_INVITATION_TYPE_EXISTING_USER

		// Validate the user ID and check if user exists
		userID, err := uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
		}

		_, err = s.userStore.FindByID(&userID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
	} else {
		// No user ID provided
		req.Type = pb.InvitationType_INVITATION_TYPE_NEW_USER
	}
	// check if email is provided
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email must be provided for invitation")
	}

	// Search for user by email
	existingUser, err := s.userStore.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		// User found, set type to existing user and set the user ID
		req.Type = pb.InvitationType_INVITATION_TYPE_EXISTING_USER
		req.UserId = existingUser.ID.String()

		if contextData.OrganisationID == existingUser.CurrentOrganisationID {
			// If the user is already part of the organization, we can return an error
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists in the organization", req.Email)
		}

		req.UserId = existingUser.ID.String() // Set the user ID from the found user
		req.RoleIds = []string{}              // Reset role IDs since we will assign them based on the user
	} else {
		// No existing user found with this email, set type to new user
		req.Type = pb.InvitationType_INVITATION_TYPE_NEW_USER
	}

	// Set default expiration if not provided
	expirationHours := req.ExpirationHours
	if expirationHours == 0 {
		expirationHours = 72 // Default 72 hours
	}

	deptId, err := uuid.Parse(req.DepartmentId)
	if err != nil {
		deptId = uuid.Nil // Default to nil if parsing fails
	}
	teamId, err := uuid.Parse(req.TeamId)
	if err != nil {
		teamId = uuid.Nil // Default to nil if parsing fails
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		userId = uuid.Nil // Default to nil if parsing fails
	}

	// give default roles for the organisation in addition to the invitation roles
	defaultRole, err := s.rbacStore.FindDefaultRole(contextData.OrganisationID)
	if err != nil {
		log.Printf("Error while fetching default role: %v", err)
	} else {
		req.RoleIds = append(req.RoleIds, defaultRole.ID.String())
	}

	// Create invitation
	inv := NewInvitation(
		req.Email,
		contextData.UserID,
		InvitationType(req.Type),
		contextData.OrganisationID,
		contextData.BranchID,
		deptId,
		teamId,
		userId,
		req.RoleIds,
		expirationHours,
	)

	// Save to database
	err = s.repo.Create(ctx, inv)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create invitation: %v", err)
	}

	// Send invitation email
	if err := s.sendInvitationEmail(inv, org); err != nil {
		log.Printf("Failed to send invitation email: %v", err)
		// We can still return the invitation even if email sending fails
	}

	return inv.ToProto(), nil
}

// GetInvitation retrieves an invitation by ID
func (s *InvitationServer) GetInvitation(ctx context.Context, req *pb.GetInviteRequest) (*pb.Invitation, error) {
	log.Println("Get Invitation")

	// Check service-level authentication
	contextData, err := s.authStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(codes.PermissionDenied, "forbidden: you do not have access to get invitation resource")
	}

	inv, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get invitation: %v", err)
	}
	if inv == nil {
		return nil, status.Error(codes.NotFound, "invitation not found")
	}

	return inv.ToProto(), nil
}

// ListInvitations retrieves a paginated list of invitations
func (s *InvitationServer) ListInvitations(ctx context.Context, req *pb.ListInvitationsRequest) (*pb.ListInvitationsResponse, error) {
	log.Println("List Invitations")

	// Check service-level authentication
	contextData, err := s.authStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(codes.PermissionDenied, "forbidden: you do not have access to list invitation resource")
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	invitations, total, err := s.repo.List(ctx, req.OrganisationId, InvitationStatus(req.Status), 1, int(pageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list invitations: %v", err)
	}

	protoInvitations := make([]*pb.Invitation, len(invitations))
	for i, inv := range invitations {
		protoInvitations[i] = inv.ToProto()
	}

	return &pb.ListInvitationsResponse{
		Invitations:   protoInvitations,
		TotalCount:    int32(total),
		NextPageToken: "",
	}, nil
}

// AcceptInvitation accepts an invitation
func (s *InvitationServer) AcceptInvitation(ctx context.Context, req *pb.AcceptInviteRequest) (*pb.Invitation, error) {
	log.Println("Accept Invitation")
	// Get invitation by ID
	inv, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get invitation: %v", err)
	}
	if inv == nil {
		return nil, status.Error(codes.NotFound, "invitation not found")
	}

	// Check if invitation is expired
	if inv.IsExpired() {
		inv.MarkAsExpired()
		err = s.repo.Update(ctx, inv)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update expired invitation: %v", err)
		}
		return nil, status.Error(codes.FailedPrecondition, "invitation has expired")
	}

	// Check if invitation is already accepted or rejected
	if inv.Status != InvitationStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "invitation is not pending")
	}

	// Handle based on invitation type
	if inv.Type == InvitationTypeNewUser {
		// For new users, we just mark the invitation as accepted
		// The actual user creation will happen when they sign up
		// For a new user, create the user account first
		newUser := &user.User{
			ID:                    uuid.New(),
			SerialNumber:          utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["users"], uuid.New().String()),
			Email:                 inv.Email,
			CurrentOrganisationID: inv.OrganisationID,
			CurrentBranchID:       inv.BranchID,
			Status:                k.USER_STATUSES()["NEW"],
			CreatedBy:             "USERS",
			OwnerType:             authctx.ORGANISATIONS,
			OwnerID:               inv.OrganisationID,
		}

		// Generate a random password
		password, err := utils.GENERATE_RANDOM_KEY(10)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate password: %v", err)
		}
		newUser.HashedPassword = utils.HASH_PASSWORD(password)
		// Add roles from invitation
		for _, roleID := range inv.RoleIDs {
			role, err := s.rbacStore.FindRoleByID(roleID)
			if err != nil {
				log.Printf("Error finding role %s: %v", roleID, err)
				continue
			}
			newUser.Roles = append(newUser.Roles, *role)
		}

		// If no roles specified, assign default role
		if len(newUser.Roles) == 0 {
			defaultRole, err := s.rbacStore.FindDefaultRole(inv.OrganisationID)
			if err != nil {
				log.Printf("Error while fetching default role: %v", err)
			} else {
				newUser.Roles = append(newUser.Roles, *defaultRole)
			}
		}
		// Save the new user
		if err := newUser.CREATE_AGENT_STATUS(s.agentStatusStore); err != nil {
			return nil, status.Error(500, "error while creating agent status")
		}
		savedUser, err := s.userStore.Save(newUser, true)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to save user: %v", err)
		}

		// Update invitation with the new user ID
		inv.UserID = savedUser.ID
		inv.Accept()
		err = s.repo.Update(ctx, inv)
		// Send welcome email
		if emailErr := s.emailService.SEND_WELCOME_EMAIL(savedUser.Email, savedUser.Username, savedUser.FirstName, password); emailErr != nil {
			log.Printf("Error sending welcome email: %v", emailErr)
		}
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update invitation: %v", err)
		}
	} else {
		// For existing users, we need to add them to the organization structure
		// and assign roles
		err = s.addUserToOrganization(inv)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to add user to organization: %v", err)
		}

		// Mark invitation as accepted
		inv.Accept()
		err = s.repo.Update(ctx, inv)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update invitation: %v", err)
		}
	}

	return inv.ToProto(), nil
}

// RejectInvitation rejects an invitation
func (s *InvitationServer) RejectInvitation(ctx context.Context, req *pb.RejectInviteRequest) (*pb.Invitation, error) {
	log.Println("Reject Invitation")

	// Check service-level authentication
	contextData, err := s.authStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(codes.PermissionDenied, "forbidden: you do not have access to reject invitation resource")
	}

	// Get invitation by ID
	inv, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get invitation: %v", err)
	}
	if inv == nil {
		return nil, status.Error(codes.NotFound, "invitation not found")
	}

	// Check if invitation is expired
	if inv.IsExpired() {
		inv.MarkAsExpired()
		err = s.repo.Update(ctx, inv)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update expired invitation: %v", err)
		}
		return nil, status.Error(codes.FailedPrecondition, "invitation has expired")
	}

	// Check if invitation is already accepted or rejected
	if inv.Status != InvitationStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "invitation is not pending")
	}

	// Mark invitation as rejected
	inv.Reject()
	err = s.repo.Update(ctx, inv)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update invitation: %v", err)
	}

	return inv.ToProto(), nil
}

// CancelInvitation cancels an invitation
func (s *InvitationServer) CancelInvitation(ctx context.Context, req *pb.CancelInviteRequest) (*pb.Invitation, error) {
	log.Println("Cancel Invitation")

	// Check service-level authentication
	contextData, err := s.authStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(codes.PermissionDenied, "forbidden: you do not have access to cancel invitation resource")
	}

	// Get invitation by ID
	inv, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get invitation: %v", err)
	}
	if inv == nil {
		return nil, status.Error(codes.NotFound, "invitation not found")
	}

	// Check if invitation is already accepted or rejected
	if inv.Status != InvitationStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "invitation is not pending")
	}

	// Cancel invitation
	inv.Cancel()
	err = s.repo.Update(ctx, inv)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update invitation: %v", err)
	}

	return inv.ToProto(), nil
}

// Helper function to send invitation email
func (s *InvitationServer) sendInvitationEmail(inv *Invitation, org *organisation.Organisation) error {
	// Build invitation URL
	invitationURL := fmt.Sprintf("%s/invitations/%s?token=%s", s.baseURL, inv.ID, inv.Token)

	// Get expiration hours
	expirationHours := int(time.Until(inv.ExpiresAt).Hours())

	// Generate email content using the template
	body, err := templates.INVITATION_EMAIL(templates.InvitationEmailData{
		OrganisationName: org.OrganisationName,
		InvitationURL:    invitationURL,
		ExpirationHours:  expirationHours,
		ClientEmail:      inv.Email,
		SupportEmail:     "support@kyla.com",
		Year:             fmt.Sprintf("%d", time.Now().Year()),
	})
	if err != nil {
		return fmt.Errorf("failed to generate email template: %v", err)
	}

	err = s.emailService.SEND_INVITATION_EMAIL(inv.Email, body)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// Helper function to add user to organization
func (s *InvitationServer) addUserToOrganization(inv *Invitation) error {
	// Add user to organization
	if err := s.organisationStore.DB.Model(&organisation.Organisation{ID: inv.OrganisationID}).Association("Users").Append(&user.User{ID: inv.UserID}); err != nil {
		return fmt.Errorf("failed to add user to organization: %v", err)
	}

	// Add user to branch if specified
	if inv.BranchID != uuid.Nil {
		if err := s.branchStore.DB.Model(&branch.Branch{ID: inv.BranchID}).Association("Users").Append(&user.User{ID: inv.UserID}); err != nil {
			return fmt.Errorf("failed to add user to branch: %v", err)
		}
	}

	// Add user to department if specified
	if inv.DepartmentID != uuid.Nil {
		if err := s.departmentStore.DB.Model(&department.Department{ID: inv.DepartmentID}).Association("Users").Append(&user.User{ID: inv.UserID}); err != nil {
			return fmt.Errorf("failed to add user to department: %v", err)
		}
	}

	// Add user to team if specified
	if inv.TeamID != uuid.Nil {
		if err := s.teamStore.DB.Model(&team.Team{ID: inv.TeamID}).Association("Users").Append(&user.User{ID: inv.UserID}); err != nil {
			return fmt.Errorf("failed to add user to team: %v", err)
		}
	}

	// Assign roles to user
	for _, roleID := range inv.RoleIDs {
		if err := s.rbacStore.DB.Model(&user.User{ID: inv.UserID}).Association("Roles").Append(&rbac.Role{ID: uuid.MustParse(roleID)}); err != nil {
			return fmt.Errorf("failed to assign role to user: %v", err)
		}
	}

	return nil
}
