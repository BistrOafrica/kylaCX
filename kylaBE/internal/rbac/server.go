package rbac

import (
	"context"
	"fmt"
	"kyla-be/internal/authctx"
	casbinsvc "kyla-be/internal/casbin"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// AuthGateway abstracts auth-metadata resolution needed by RbacServer.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	GetUserRequestMetadata(ctx context.Context, dataChan chan *authctx.RequestMetadata, errChan chan error)
}

// UserQuerier abstracts user lookups needed by RbacServer without importing the user package.
type UserQuerier interface {
	// FindUsersByRoleIds returns the proto representation of users assigned to the given role IDs.
	FindUsersByRoleIds(roleIds []string) ([]*pb.User, error)
	// FindRolesByUserID returns all roles belonging to the given user.
	FindRolesByUserID(id *uuid.UUID) ([]*Role, error)
}

// --- minimal local proxy structs for raw branch/dept/team GORM queries ---

type rbacBranch struct {
	ID        uuid.UUID `gorm:"primarykey"`
	Name      string
	OwnerID   uuid.UUID
	OwnerType string
}

func (rbacBranch) TableName() string { return "branches" }

type rbacDepartment struct {
	ID             uuid.UUID `gorm:"primarykey"`
	DepartmentName string
	OwnerID        uuid.UUID
	OwnerType      string
}

func (rbacDepartment) TableName() string { return "departments" }

type rbacTeam struct {
	ID        uuid.UUID `gorm:"primarykey"`
	Name      string
	OwnerID   uuid.UUID
	OwnerType string
}

func (rbacTeam) TableName() string { return "teams" }

// RbacServer implements the Role and Permission gRPC services.
type RbacServer struct {
	pb.UnimplementedRoleServiceServer
	pb.UnimplementedPermissionServiceServer
	AuthGateway AuthGateway
	rbacStore   *RbacStore
	UserStore   UserQuerier
	// DB is used for cross-domain raw queries (branch/dept/team scoping).
	DB       *gorm.DB
	enforcer *casbinsvc.Enforcer // may be nil if Casbin is not wired
}

// NewRbacServer constructs a new RbacServer.
func NewRbacServer(
	rbacStore *RbacStore,
	authGateway AuthGateway,
	userStore UserQuerier,
	db *gorm.DB,
	enforcer *casbinsvc.Enforcer,
) *RbacServer {
	return &RbacServer{
		rbacStore:   rbacStore,
		AuthGateway: authGateway,
		UserStore:   userStore,
		DB:          db,
		enforcer:    enforcer,
	}
}

// CreateRole handles the CreateRole gRPC call.
func (r *RbacServer) CreateRole(ctx context.Context, request *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	log.Println("Create Role")
	contextData := &authctx.RequestMetadata{}
	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go r.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	role := PbRoleToRole(request.GetRole())
	role.ID = uuid.New()
	role.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], role.ID.String())
	role.CreatedBy = contextData.UserID.String()
	role.UpdatedBy = contextData.UserID.String()

	newRole, err := r.rbacStore.SaveRole(role)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to create role")
	}

	return &pb.CreateRoleResponse{
		Role: RoleToPbRole(newRole),
	}, nil
}

// ReadRole handles the ReadRole gRPC call.
func (r *RbacServer) ReadRole(ctx context.Context, request *pb.ReadRoleRequest) (*pb.ReadRoleResponse, error) {
	log.Println("Read Role")

	reqAuth, authErr := r.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil {
		return nil, status.Error(403, "Forbidden: You don't have permission to read a role")
	}
	if reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to read a role")
	}

	roleID := request.GetId()
	role, err := r.rbacStore.FindRoleByID(roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get role")
	}

	return &pb.ReadRoleResponse{
		Role: RoleToPbRole(role),
	}, nil
}

// ReadRoles handles the ReadRoles gRPC call, aggregating roles across the org hierarchy.
func (r *RbacServer) ReadRoles(ctx context.Context, request *pb.ReadRolesRequest) (*pb.ReadRolesResponse, error) {
	log.Println("Read Roles")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go r.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextData *authctx.RequestMetadata
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get roles")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	allRoles := []*Role{}

	// 1. Organisation-level roles
	orgScope := &authctx.OpScope{
		Owner: authctx.ORGANISATIONS,
		ID:    contextData.OrganisationID.String(),
	}
	orgRoles, err := r.rbacStore.FindAllByScope(orgScope)
	if err != nil {
		log.Printf("Error fetching organisation roles: %v", err)
	} else {
		allRoles = append(allRoles, orgRoles...)
		log.Printf("Found %d organisation roles", len(orgRoles))
	}

	// 2. Branches in the organisation and their roles
	var branches []*rbacBranch
	err = r.DB.Where("owner_id = ? AND owner_type = ?", contextData.OrganisationID.String(), string(authctx.ORGANISATIONS)).Find(&branches).Error
	if err != nil {
		log.Printf("Error fetching branches: %v", err)
	} else {
		log.Printf("Found %d branches in organisation", len(branches))
		for _, b := range branches {
			branchScope := &authctx.OpScope{
				Owner: authctx.BRANCHES,
				ID:    b.ID.String(),
			}
			branchRoles, err := r.rbacStore.FindAllByScope(branchScope)
			if err != nil {
				log.Printf("Error fetching roles for branch %s: %v", b.ID, err)
			} else {
				allRoles = append(allRoles, branchRoles...)
				log.Printf("Found %d roles for branch %s", len(branchRoles), b.Name)
			}
		}
	}

	// 3. Organisation-level departments and their roles
	var orgDepartments []*rbacDepartment
	err = r.DB.Where("owner_id = ? AND owner_type = ?", contextData.OrganisationID.String(), string(authctx.ORGANISATIONS)).Find(&orgDepartments).Error
	if err != nil {
		log.Printf("Error fetching organisation departments: %v", err)
	} else {
		log.Printf("Found %d departments in organisation", len(orgDepartments))
		for _, dept := range orgDepartments {
			deptScope := &authctx.OpScope{
				Owner: authctx.DEPARTMENTS,
				ID:    dept.ID.String(),
			}
			deptRoles, err := r.rbacStore.FindAllByScope(deptScope)
			if err != nil {
				log.Printf("Error fetching roles for organisation department %s: %v", dept.ID, err)
			} else {
				allRoles = append(allRoles, deptRoles...)
				log.Printf("Found %d roles for organisation department %s", len(deptRoles), dept.DepartmentName)
			}
		}
	}

	// 4. Organisation-level teams and their roles
	var orgTeams []*rbacTeam
	err = r.DB.Where("owner_id = ? AND owner_type = ?", contextData.OrganisationID.String(), string(authctx.ORGANISATIONS)).Find(&orgTeams).Error
	if err != nil {
		log.Printf("Error fetching organisation teams: %v", err)
	} else {
		log.Printf("Found %d teams in organisation", len(orgTeams))
		for _, t := range orgTeams {
			teamScope := &authctx.OpScope{
				Owner: authctx.TEAMS,
				ID:    t.ID.String(),
			}
			teamRoles, err := r.rbacStore.FindAllByScope(teamScope)
			if err != nil {
				log.Printf("Error fetching roles for organisation team %s: %v", t.ID, err)
			} else {
				allRoles = append(allRoles, teamRoles...)
				log.Printf("Found %d roles for organisation team %s", len(teamRoles), t.Name)
			}
		}
	}

	// 5. Teams belonging to branches
	for _, b := range branches {
		var branchTeams []*rbacTeam
		err = r.DB.Where("owner_id = ? AND owner_type = ?", b.ID.String(), string(authctx.BRANCHES)).Find(&branchTeams).Error
		if err != nil {
			log.Printf("Error fetching teams for branch %s: %v", b.ID, err)
		} else {
			log.Printf("Found %d teams in branch %s", len(branchTeams), b.Name)
			for _, t := range branchTeams {
				teamScope := &authctx.OpScope{
					Owner: authctx.TEAMS,
					ID:    t.ID.String(),
				}
				teamRoles, err := r.rbacStore.FindAllByScope(teamScope)
				if err != nil {
					log.Printf("Error fetching roles for branch team %s: %v", t.ID, err)
				} else {
					allRoles = append(allRoles, teamRoles...)
					log.Printf("Found %d roles for branch team %s", len(teamRoles), t.Name)
				}
			}
		}
	}

	// 6. Departments belonging to branches
	for _, b := range branches {
		var branchDepartments []*rbacDepartment
		err = r.DB.Where("owner_id = ? AND owner_type = ?", b.ID.String(), string(authctx.BRANCHES)).Find(&branchDepartments).Error
		if err != nil {
			log.Printf("Error fetching departments for branch %s: %v", b.ID, err)
		} else {
			log.Printf("Found %d departments in branch %s", len(branchDepartments), b.Name)
			for _, dept := range branchDepartments {
				deptScope := &authctx.OpScope{
					Owner: authctx.DEPARTMENTS,
					ID:    dept.ID.String(),
				}
				deptRoles, err := r.rbacStore.FindAllByScope(deptScope)
				if err != nil {
					log.Printf("Error fetching roles for branch department %s: %v", dept.ID, err)
				} else {
					allRoles = append(allRoles, deptRoles...)
					log.Printf("Found %d roles for branch department %s", len(deptRoles), dept.DepartmentName)
				}
			}
		}
	}

	// 7. Teams belonging to departments (all departments: org-level + branch-level)
	allDepartments := append(orgDepartments, []*rbacDepartment{}...)
	for _, b := range branches {
		var branchDepts []*rbacDepartment
		err = r.DB.Where("owner_id = ? AND owner_type = ?", b.ID.String(), string(authctx.BRANCHES)).Find(&branchDepts).Error
		if err == nil {
			allDepartments = append(allDepartments, branchDepts...)
		}
	}
	for _, dept := range allDepartments {
		var deptTeams []*rbacTeam
		err = r.DB.Where("owner_id = ? AND owner_type = ?", dept.ID.String(), string(authctx.DEPARTMENTS)).Find(&deptTeams).Error
		if err != nil {
			log.Printf("Error fetching teams for department %s: %v", dept.ID, err)
		} else {
			log.Printf("Found %d teams in department %s", len(deptTeams), dept.DepartmentName)
			for _, t := range deptTeams {
				teamScope := &authctx.OpScope{
					Owner: authctx.TEAMS,
					ID:    t.ID.String(),
				}
				teamRoles, err := r.rbacStore.FindAllByScope(teamScope)
				if err != nil {
					log.Printf("Error fetching roles for department team %s: %v", t.ID, err)
				} else {
					allRoles = append(allRoles, teamRoles...)
					log.Printf("Found %d roles for department team %s", len(teamRoles), t.Name)
				}
			}
		}
	}

	log.Printf("Total roles aggregated: %d", len(allRoles))

	return &pb.ReadRolesResponse{
		Roles: RolesToPbRoles(allRoles),
	}, nil
}

// UpdateRole handles the UpdateRole gRPC call.
func (r *RbacServer) UpdateRole(ctx context.Context, request *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	log.Println("Update Role")

	reqAuth, authErr := r.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to update a role")
	}

	if err := r.rbacStore.UpdateRole(PbRoleToRole(request.GetRole())); err != nil {
		return nil, status.Error(codes.Internal, "Failed to update role")
	}

	return &pb.UpdateRoleResponse{
		Role: request.GetRole(),
	}, nil
}

// DeleteRole handles the DeleteRole gRPC call.
func (r *RbacServer) DeleteRole(ctx context.Context, request *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	log.Println("Delete Role")

	reqAuth, authErr := r.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to delete a role")
	}

	if err := r.rbacStore.DeleteRole(request.GetId()); err != nil {
		return nil, status.Error(codes.Internal, "Failed to delete role")
	}
	return &pb.DeleteRoleResponse{
		Success: true,
	}, nil
}

// ReadPermissions handles the ReadPermissions gRPC call.
func (r *RbacServer) ReadPermissions(ctx context.Context, request *pb.ReadPermissionsRequest) (*pb.ReadPermissionsResponse, error) {
	log.Println("Read Permissions")

	reqAuth, authErr := r.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to read permissions")
	}
	return &pb.ReadPermissionsResponse{
		Permissions: ConvertMapToPbPermissions(k.ALL_PERMISSIONS()),
	}, nil
}

func (r *RbacServer) getPermissionRoles(permission string, scope *authctx.OpScope, rolesChan chan []string, errChan chan error) {
	roles, err := r.rbacStore.FindRolesByPermission(permission, scope)
	if err != nil {
		errChan <- err
		return
	}
	roleIds := []string{}
	for _, role := range roles {
		roleIds = append(roleIds, role.ID.String())
	}
	rolesChan <- roleIds
}

func (r *RbacServer) getRoleUsers(roleIds []string, usersChan chan []*pb.User, errChan chan error) {
	users, err := r.UserStore.FindUsersByRoleIds(roleIds)
	if err != nil {
		errChan <- err
		return
	}
	usersChan <- users
}

// AddPermissionToRole handles the AddPermissionToRole gRPC call.
func (r *RbacServer) AddPermissionToRole(ctx context.Context, req *pb.PermissionToRoleRequest) (*pb.PermissionToRoleResponse, error) {
	log.Println("Add Permission To Role")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go r.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	role, err := r.rbacStore.AddPermissionToRole(req.GetRoleId(), req.GetPermissionCodenames())
	if err != nil {
		return nil, status.Errorf(500, "error while adding permission to role %v", err)
	}

	// Sync new permissions to Casbin (non-blocking — failures are logged, not fatal).
	if r.enforcer != nil && role != nil {
		go r.syncRolePolicies(role, req.GetPermissionCodenames(), true)
	}

	return &pb.PermissionToRoleResponse{
		Role: RoleToPbRole(role),
	}, nil
}

// RemovePermissionFromRole handles the RemovePermissionFromRole gRPC call.
func (r *RbacServer) RemovePermissionFromRole(ctx context.Context, req *pb.PermissionToRoleRequest) (*pb.PermissionToRoleResponse, error) {
	log.Println("Remove Permission From Role")

	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go r.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	role, err := r.rbacStore.RemovePermissionFromRole(req.GetRoleId(), req.GetPermissionCodenames())
	if err != nil {
		return nil, status.Errorf(500, "error while removing permission from role %v", err)
	}

	// Remove synced Casbin policies (non-blocking — failures are logged, not fatal).
	if r.enforcer != nil && role != nil {
		go r.syncRolePolicies(role, req.GetPermissionCodenames(), false)
	}

	return &pb.PermissionToRoleResponse{
		Role: RoleToPbRole(role),
	}, nil
}

// syncRolePolicies adds (add=true) or removes (add=false) Casbin policies for codenames.
// Runs in a goroutine — errors are logged but do not surface to the caller.
func (r *RbacServer) syncRolePolicies(role *Role, codeNames []string, add bool) {
	if r.enforcer == nil || role == nil {
		return
	}
	// Derive the Casbin domain from the role's owner type.
	var domain string
	switch role.OwnerType {
	case "ORGANISATIONS":
		domain = fmt.Sprintf("org:%s", role.OwnerID.String())
	case "BRANCHES":
		domain = fmt.Sprintf("branch:%s", role.OwnerID.String())
	case "TEAMS":
		domain = fmt.Sprintf("team:%s", role.OwnerID.String())
	case "DEPARTMENTS":
		domain = fmt.Sprintf("dept:%s", role.OwnerID.String())
	default:
		return // unknown owner type — skip
	}

	subject := fmt.Sprintf("role:%s", role.ID.String())
	for _, codeName := range codeNames {
		resource, action, ok := casbinsvc.CodeNameToPolicy(codeName)
		if !ok {
			continue
		}
		var err error
		if add {
			err = r.enforcer.AddPolicy(subject, domain, resource, action)
		} else {
			err = r.enforcer.RemovePolicy(subject, domain, resource, action)
		}
		if err != nil {
			log.Printf("[rbac] casbin policy sync error for codename %s: %v", codeName, err)
		}
	}
}

// ReadUserRoles handles the ReadUserRoles gRPC call.
func (r *RbacServer) ReadUserRoles(ctx context.Context, req *pb.ReadRolesRequest) (*pb.ReadRolesResponse, error) {
	log.Println("Read User Roles")
	contextData := &authctx.RequestMetadata{}
	contextDataChan := make(chan *authctx.RequestMetadata)
	errChan := make(chan error)
	go r.AuthGateway.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	roles, err := r.UserStore.FindRolesByUserID(&contextData.UserID)
	if err != nil {
		return nil, status.Errorf(500, "error while fetching user roles %v", err)
	}

	return &pb.ReadRolesResponse{
		Roles: RolesToPbRoles(roles),
	}, nil
}

// ReadUsersByPermission handles the ReadUsersByPermission gRPC call.
func (r *RbacServer) ReadUsersByPermission(_ context.Context, request *pb.ReadUsersByPermissionRequest) (*pb.ReadUsersByPermissionResponse, error) {
	log.Println("Read Users By Permission")
	scope := authctx.PbScopeToOpScope(request.GetScope())
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
	case err := <-errChan:
		return nil, status.Errorf(codes.Internal, "Failed to get roles by permission %v", err)
	}

	return &pb.ReadUsersByPermissionResponse{
		Users: users,
	}, nil
}
