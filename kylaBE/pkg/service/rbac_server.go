package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RbacServer struct {
	pb.UnimplementedRoleServiceServer
	pb.UnimplementedPermissionServiceServer
	AuthStore       *AuthStore
	rbacStore       *RbacStore
	UserStore       *UserStore
	BranchStore     *BranchStore
	DepartmentStore *DepartmentStore
	TeamStore       *TeamStore
}

func NewRbacServer(rbacStore *RbacStore, AuthStore *AuthStore, UserStore *UserStore, BranchStore *BranchStore, DepartmentStore *DepartmentStore, TeamStore *TeamStore) *RbacServer {
	return &RbacServer{
		rbacStore:       rbacStore,
		AuthStore:       AuthStore,
		UserStore:       UserStore,
		BranchStore:     BranchStore,
		DepartmentStore: DepartmentStore,
		TeamStore:       TeamStore,
	}
}

// Roles
func (r *RbacServer) CreateRole(ctx context.Context, request *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	log.Println("Create Role")
	contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go r.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
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

func (r *RbacServer) ReadRole(ctx context.Context, request *pb.ReadRoleRequest) (*pb.ReadRoleResponse, error) {
	log.Println("Read Role")

	reqAuth, authErr := r.AuthStore.GetServiceAuthMetadata(ctx)
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

func (r *RbacServer) ReadRoles(ctx context.Context, request *pb.ReadRolesRequest) (*pb.ReadRolesResponse, error) {
	log.Println("Read Roles")

	// Get user authentication and context data
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go r.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextData *RequestMetadata
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get roles")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	// Collect all roles from the organizational hierarchy
	allRoles := []*Role{}

	// 1. Get organisation-level roles
	orgScope := &OpScope{
		Owner: OwnerType(ORGANISATIONS),
		ID:    contextData.OrganisationID.String(),
	}
	orgRoles, err := r.rbacStore.FindAllByScope(orgScope)
	if err != nil {
		log.Printf("Error fetching organisation roles: %v", err)
	} else {
		allRoles = append(allRoles, orgRoles...)
		log.Printf("Found %d organisation roles", len(orgRoles))
	}

	// 2. Get all branches in the organisation and their roles
	var branches []*Branch
	err = r.BranchStore.DB.Where("owner_id = ? AND owner_type = ?", contextData.OrganisationID.String(), string(ORGANISATIONS)).Find(&branches).Error
	if err != nil {
		log.Printf("Error fetching branches: %v", err)
	} else {
		log.Printf("Found %d branches in organisation", len(branches))
		for _, branch := range branches {
			branchScope := &OpScope{
				Owner: OwnerType(BRANCHES),
				ID:    branch.ID.String(),
			}
			branchRoles, err := r.rbacStore.FindAllByScope(branchScope)
			if err != nil {
				log.Printf("Error fetching roles for branch %s: %v", branch.ID, err)
			} else {
				allRoles = append(allRoles, branchRoles...)
				log.Printf("Found %d roles for branch %s", len(branchRoles), branch.Name)
			}
		}
	}

	// 3. Get organisation department roles (departments that belong directly to organisation)
	var orgDepartments []*Department
	err = r.DepartmentStore.db.Where("owner_id = ? AND owner_type = ?", contextData.OrganisationID.String(), string(ORGANISATIONS)).Find(&orgDepartments).Error
	if err != nil {
		log.Printf("Error fetching organisation departments: %v", err)
	} else {
		log.Printf("Found %d departments in organisation", len(orgDepartments))
		for _, department := range orgDepartments {
			deptScope := &OpScope{
				Owner: OwnerType(DEPARTMENTS),
				ID:    department.ID.String(),
			}
			deptRoles, err := r.rbacStore.FindAllByScope(deptScope)
			if err != nil {
				log.Printf("Error fetching roles for organisation department %s: %v", department.ID, err)
			} else {
				allRoles = append(allRoles, deptRoles...)
				log.Printf("Found %d roles for organisation department %s", len(deptRoles), department.DepartmentName)
			}
		}
	}

	// 4. Get organisation team roles (teams that belong directly to organisation)
	var orgTeams []*Team
	err = r.TeamStore.DB.Where("owner_id = ? AND owner_type = ?", contextData.OrganisationID.String(), string(ORGANISATIONS)).Find(&orgTeams).Error
	if err != nil {
		log.Printf("Error fetching organisation teams: %v", err)
	} else {
		log.Printf("Found %d teams in organisation", len(orgTeams))
		for _, team := range orgTeams {
			teamScope := &OpScope{
				Owner: OwnerType(TEAMS),
				ID:    team.ID.String(),
			}
			teamRoles, err := r.rbacStore.FindAllByScope(teamScope)
			if err != nil {
				log.Printf("Error fetching roles for organisation team %s: %v", team.ID, err)
			} else {
				allRoles = append(allRoles, teamRoles...)
				log.Printf("Found %d roles for organisation team %s", len(teamRoles), team.Name)
			}
		}
	}

	// 5. Get branch team roles (teams that belong to branches)
	for _, branch := range branches {
		var branchTeams []*Team
		err = r.TeamStore.DB.Where("owner_id = ? AND owner_type = ?", branch.ID.String(), string(BRANCHES)).Find(&branchTeams).Error
		if err != nil {
			log.Printf("Error fetching teams for branch %s: %v", branch.ID, err)
		} else {
			log.Printf("Found %d teams in branch %s", len(branchTeams), branch.Name)
			for _, team := range branchTeams {
				teamScope := &OpScope{
					Owner: OwnerType(TEAMS),
					ID:    team.ID.String(),
				}
				teamRoles, err := r.rbacStore.FindAllByScope(teamScope)
				if err != nil {
					log.Printf("Error fetching roles for branch team %s: %v", team.ID, err)
				} else {
					allRoles = append(allRoles, teamRoles...)
					log.Printf("Found %d roles for branch team %s", len(teamRoles), team.Name)
				}
			}
		}
	}

	// 6. Get branch department roles (departments that belong to branches)
	for _, branch := range branches {
		var branchDepartments []*Department
		err = r.DepartmentStore.db.Where("owner_id = ? AND owner_type = ?", branch.ID.String(), string(BRANCHES)).Find(&branchDepartments).Error
		if err != nil {
			log.Printf("Error fetching departments for branch %s: %v", branch.ID, err)
		} else {
			log.Printf("Found %d departments in branch %s", len(branchDepartments), branch.Name)
			for _, department := range branchDepartments {
				deptScope := &OpScope{
					Owner: OwnerType(DEPARTMENTS),
					ID:    department.ID.String(),
				}
				deptRoles, err := r.rbacStore.FindAllByScope(deptScope)
				if err != nil {
					log.Printf("Error fetching roles for branch department %s: %v", department.ID, err)
				} else {
					allRoles = append(allRoles, deptRoles...)
					log.Printf("Found %d roles for branch department %s", len(deptRoles), department.DepartmentName)
				}
			}
		}
	}

	// 7. Get department team roles (teams that belong to departments)
	// First get all departments (both org and branch departments)
	allDepartments := append(orgDepartments, []*Department{}...)
	for _, branch := range branches {
		var branchDepts []*Department
		err = r.DepartmentStore.db.Where("owner_id = ? AND owner_type = ?", branch.ID.String(), string(BRANCHES)).Find(&branchDepts).Error
		if err == nil {
			allDepartments = append(allDepartments, branchDepts...)
		}
	}

	for _, department := range allDepartments {
		var deptTeams []*Team
		err = r.TeamStore.DB.Where("owner_id = ? AND owner_type = ?", department.ID.String(), string(DEPARTMENTS)).Find(&deptTeams).Error
		if err != nil {
			log.Printf("Error fetching teams for department %s: %v", department.ID, err)
		} else {
			log.Printf("Found %d teams in department %s", len(deptTeams), department.DepartmentName)
			for _, team := range deptTeams {
				teamScope := &OpScope{
					Owner: OwnerType(TEAMS),
					ID:    team.ID.String(),
				}
				teamRoles, err := r.rbacStore.FindAllByScope(teamScope)
				if err != nil {
					log.Printf("Error fetching roles for department team %s: %v", team.ID, err)
				} else {
					allRoles = append(allRoles, teamRoles...)
					log.Printf("Found %d roles for department team %s", len(teamRoles), team.Name)
				}
			}
		}
	}

	log.Printf("Total roles aggregated: %d", len(allRoles))

	return &pb.ReadRolesResponse{
		Roles: RolesToPbRoles(allRoles),
	}, nil
}

func (r *RbacServer) UpdateRole(ctx context.Context, request *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	log.Println("Update Role")

	reqAuth, authErr := r.AuthStore.GetServiceAuthMetadata(ctx)
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

func (r *RbacServer) DeleteRole(ctx context.Context, request *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	log.Println("Delete Role")

	reqAuth, authErr := r.AuthStore.GetServiceAuthMetadata(ctx)
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

// PERMISSIONS

func (r *RbacServer) ReadPermissions(ctx context.Context, request *pb.ReadPermissionsRequest) (*pb.ReadPermissionsResponse, error) {
	log.Println("Read Permissions")

	reqAuth, authErr := r.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to read permissions")
	}
	return &pb.ReadPermissionsResponse{
		Permissions: ConvertMapToPbPermissions(k.ALL_PERMISSIONS()),
	}, nil
}

func (r *RbacServer) getPermissionRoles(permission string, scope *OpScope, rolesChan chan []string, errChan chan error) {
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
	usersChan <- UsersToPbUsers(users)
}

func (r *RbacServer) AddPermissionToRole(ctx context.Context, req *pb.PermissionToRoleRequest) (*pb.PermissionToRoleResponse, error) {
	log.Println("Add Permission To Role")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go r.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

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

	return &pb.PermissionToRoleResponse{
		Role: RoleToPbRole(role),
	}, nil
}

func (r *RbacServer) RemovePermissionFromRole(ctx context.Context, req *pb.PermissionToRoleRequest) (*pb.PermissionToRoleResponse, error) {
	log.Println("Remove Permission From Role")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go r.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

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

	return &pb.PermissionToRoleResponse{
		Role: RoleToPbRole(role),
	}, nil
}

func (r *RbacServer) ReadUserRoles(ctx context.Context, req *pb.ReadRolesRequest) (*pb.ReadRolesResponse, error) {
	log.Println("Read User Roles")
	contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go r.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	// var contextData *RequestMetadata
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	user, err := r.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Errorf(500, "error while fetching user roles %v", err)
	}

	roles := []*Role{}
	for _, role := range user.Roles {
		roles = append(roles, &role)
	}

	return &pb.ReadRolesResponse{
		Roles: RolesToPbRoles(roles),
	}, nil
}
