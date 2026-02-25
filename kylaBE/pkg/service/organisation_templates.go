package service

import (
	"context"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

// ReadTemplates returns all organization templates, optionally filtered by industry
func (o *Org) ReadTemplates(ctx context.Context, request *pb.ReadTemplatesRequest) (*pb.ReadTemplatesResponse, error) {
	log.Println("Read Templates")

	// Verify auth
	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to read templates")
	}

	// Get templates from k package
	result := &pb.ReadTemplatesResponse{
		IndustryTemplates: []*pb.IndustryTemplates{},
	}

	// Convert templates to protobuf format
	for industry, subIndustryMap := range k.OrganizationTemplates {
		// If industry filter is specified and doesn't match, skip
		if request.Industry != "" && request.Industry != industry {
			continue
		}
		if industry == "Default" {
			continue
		}

		industryTemplates := &pb.IndustryTemplates{
			Industry:             industry,
			SubindustryTemplates: make(map[string]*pb.OrganizationTemplate),
		}

		for subIndustry, template := range subIndustryMap {
			// Convert template to protobuf format
			pbTemplate := &pb.OrganizationTemplate{
				Industry:            template.Industry,
				Subindustry:         template.Subindustry,
				TemplateName:        template.TemplateName,
				TemplateDescription: template.TemplateDescription,
				Branches:            []*pb.BranchTemplate{},
			}

			// Convert branches
			for _, branch := range template.Branches {
				pbBranch := &pb.BranchTemplate{
					Name:        branch.Name,
					Description: branch.Description,
					Departments: []*pb.DepartmentTemplate{},
					Teams:       []*pb.TeamTemplate{},
				}

				// Convert departments
				for _, dept := range branch.Departments {
					pbDept := &pb.DepartmentTemplate{
						Name:        dept.Name,
						Description: dept.Description,
						Teams:       []*pb.TeamTemplate{},
					}

					// Convert teams in department
					for _, team := range dept.Teams {
						pbTeam := &pb.TeamTemplate{
							Name:        team.Name,
							Description: team.Description,
						}
						pbDept.Teams = append(pbDept.Teams, pbTeam)
					}

					pbBranch.Departments = append(pbBranch.Departments, pbDept)
				}

				// Convert teams in branch
				for _, team := range branch.Teams {
					pbTeam := &pb.TeamTemplate{
						Name:        team.Name,
						Description: team.Description,
					}
					pbBranch.Teams = append(pbBranch.Teams, pbTeam)
				}

				pbTemplate.Branches = append(pbTemplate.Branches, pbBranch)
			}

			industryTemplates.SubindustryTemplates[subIndustry] = pbTemplate
		}

		result.IndustryTemplates = append(result.IndustryTemplates, industryTemplates)
	}

	return result, nil
}

// CreateDefaultStructure creates an organization structure using the default template
func (o *Org) CreateDefaultStructure(ctx context.Context, request *pb.CreateDefaultStructureRequest) (*pb.CreateDefaultStructureResponse, error) {
	log.Println("Create Default Structure")
	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to create organization structure")
	}

	// Get the organization
	orgID, err := uuid.Parse(request.OrganisationId)
	if err != nil {
		return nil, status.Error(400, "invalid organization ID")
	}

	org, err := o.OrganisationStore.FindByID(&orgID)
	if err != nil {
		return nil, status.Error(404, "organization not found")
	}

	// Get the user
	user, err := o.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Error(500, "error while fetching user")
	}
	// Get the default template
	defaultTemplate := k.OrganizationTemplates["Default"]["Generic"]
	if defaultTemplate.Industry == "" {
		return nil, status.Error(500, "default template not found")
	}

	// Call the common implementation with the default template
	result, err := o.createStructure(user, org, defaultTemplate)
	if err != nil {
		return nil, err
	}

	return &pb.CreateDefaultStructureResponse{
		Organisation: result.Organisation,
	}, nil
}

// CreateStructureWithTemplate creates an organization structure using the specified template
func (o *Org) CreateStructureWithTemplate(ctx context.Context, request *pb.CreateStructureWithTemplateRequest) (*pb.CreateStructureWithTemplateResponse, error) {
	log.Println("Create Structure With Template")
	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to create organization structure")
	}

	// Get the organization
	orgID, err := uuid.Parse(request.OrganisationId)
	if err != nil {
		return nil, status.Error(400, "invalid organization ID")
	}

	org, err := o.OrganisationStore.FindByID(&orgID)
	if err != nil {
		return nil, status.Error(404, "organization not found")
	}

	// Get the user
	user, err := o.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Error(500, "error while fetching user")
	}

	// Validate industry exists
	industryTemplates, ok := k.OrganizationTemplates[request.Industry]
	if !ok {
		return nil, status.Errorf(400, "industry '%s' not found", request.Industry)
	}

	// If subindustry is not specified, use the first one
	subindustry := request.Subindustry
	if subindustry == "" {
		// Get the first subindustry
		for sub := range industryTemplates {
			subindustry = sub
			break
		}
	}

	// Validate subindustry exists
	template, ok := industryTemplates[subindustry]
	if !ok {
		return nil, status.Errorf(400, "subindustry '%s' not found for industry '%s'", subindustry, request.Industry)
	}

	// Call the common implementation
	result, err := o.createStructure(user, org, template)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Common implementation for both CreateDefaultStructure and CreateStructureWithTemplate
func (o *Org) createStructure(user *User, org *Organisation, template k.OrganizationTemplate) (*pb.CreateStructureWithTemplateResponse, error) {

	// MARK: Branches
	for _, branchTemplate := range template.Branches {
		// Create branch
		branchID := uuid.New()
		branch := Branch{
			ID:           branchID,
			SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["branches"], branchID.String()),
			Name:         branchTemplate.Name,
			Description:  branchTemplate.Description,
			CreatedBy:    user.ID.String(),
			Status:       "ACTIVE",
			ParentID:     uuid.Nil,
			IsDefault:    false,
			OwnerType:    OwnerType(ORGANISATIONS),
			OwnerID:      org.ID,
			Users:        []User{},
			Roles:        []Role{},
			Teams:        []Team{},
		}

		// Create branch roles
		adminId := uuid.New()
		supervisorId := uuid.New()
		agentId := uuid.New()

		// MARK: Branches Roles
		branchRoles := []Role{
			{
				ID:                  adminId,
				SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], adminId.String()),
				Name:                fmt.Sprintf("%s Admin", branchTemplate.Name),
				Description:         "Branch Admin Role",
				PermissionCodeNames: k.ADMIN_PERMISSIONS(),
				CreatedBy:           user.ID.String(),
				UpdatedAt:           time.Now(),
				CreatedAt:           time.Now(),
				IsDefault:           false,
				OwnerType:           OwnerType(BRANCHES),
				OwnerID:             branchID,
			},
			{
				ID:                  supervisorId,
				SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], supervisorId.String()),
				Name:                fmt.Sprintf("%s Supervisor", branchTemplate.Name),
				Description:         "Branch Supervisor Role",
				PermissionCodeNames: k.SUPERVISOR_PERMISSIONS(),
				CreatedBy:           user.ID.String(),
				UpdatedAt:           time.Now(),
				CreatedAt:           time.Now(),
				IsDefault:           false,
				OwnerType:           OwnerType(BRANCHES),
				OwnerID:             branchID,
			},
			{
				ID:                  agentId,
				SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], agentId.String()),
				Name:                fmt.Sprintf("%s Agent", branchTemplate.Name),
				Description:         "Branch Agent Role",
				PermissionCodeNames: k.AGENT_PERMISSIONS(),
				CreatedBy:           user.ID.String(),
				UpdatedAt:           time.Now(),
				CreatedAt:           time.Now(),
				IsDefault:           false,
				OwnerType:           OwnerType(BRANCHES),
				OwnerID:             branchID,
			},
		}

		// Create branch in database first without relationships
		if err := o.BranchStore.DB.Create(&branch).Error; err != nil {
			return nil, status.Errorf(500, "failed to create branch: %v", err)
		}

		// Create roles for branch first
		for i := range branchRoles {
			if _, err := o.RoleStore.SaveRole(&branchRoles[i]); err != nil {
				return nil, status.Errorf(500, "failed to create branch role: %v", err)
			}
		}

		// MARK: Branch Departments
		for _, deptTemplate := range branchTemplate.Departments {
			deptID := uuid.New()
			dept := Department{
				ID:             deptID,
				SerialNumber:   utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["departments"], deptID.String()),
				DepartmentName: deptTemplate.Name,
				DepartmentBio:  deptTemplate.Description,
				CreatedBy:      user.ID.String(),
				Status:         "ACTIVE",
				OwnerType:      OwnerType(BRANCHES),
				OwnerID:        branchID,
				Users:          []User{*user},
				Roles:          []Role{},
			}

			// Create department roles
			deptAdminId := uuid.New()
			deptMemberId := uuid.New()

			// MARK: Department Roles
			deptRoles := []Role{
				{
					ID:                  deptAdminId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], deptAdminId.String()),
					Name:                fmt.Sprintf("%s Admin", deptTemplate.Name),
					Description:         "Department Admin Role",
					PermissionCodeNames: k.ADMIN_PERMISSIONS(),
					CreatedBy:           user.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           OwnerType(DEPARTMENTS),
					OwnerID:             deptID,
				},
				{
					ID:                  deptMemberId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], deptMemberId.String()),
					Name:                fmt.Sprintf("%s Member", deptTemplate.Name),
					Description:         "Department Member Role",
					PermissionCodeNames: k.BASIC_PERMISSIONS(),
					CreatedBy:           user.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           OwnerType(DEPARTMENTS),
					OwnerID:             deptID,
				},
			}

			// Initialize Teams slice
			dept.Teams = []Team{}

			// Save department in database first
			if err := o.BranchStore.DB.Create(&dept).Error; err != nil {
				return nil, status.Errorf(500, "failed to create department: %v", err)
			}

			// Create roles for department
			for i := range deptRoles {
				if _, err := o.RoleStore.SaveRole(&deptRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to create department role: %v", err)
				}
			}

			// MARK: Department Teams
			for _, teamTemplate := range deptTemplate.Teams {
				teamID := uuid.New()
				team := Team{
					ID:           teamID,
					SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["teams"], teamID.String()),
					Name:         teamTemplate.Name,
					Description:  teamTemplate.Description,
					CreatedBy:    user.ID.String(),
					OwnerType:    OwnerType(DEPARTMENTS),
					OwnerID:      deptID,
					Users:        []User{*user},
					Roles:        []Role{},
				}

				// Create team roles
				teamAdminId := uuid.New()
				teamMemberId := uuid.New()

				// MARK: Department Team Roles
				teamRoles := []Role{
					{
						ID:                  teamAdminId,
						SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamAdminId.String()),
						Name:                fmt.Sprintf("%s Admin", teamTemplate.Name),
						Description:         "Team Admin Role",
						PermissionCodeNames: k.ADMIN_PERMISSIONS(),
						CreatedBy:           user.ID.String(),
						UpdatedAt:           time.Now(),
						CreatedAt:           time.Now(),
						IsDefault:           false,
						OwnerType:           OwnerType(TEAMS),
						OwnerID:             teamID,
					},
					{
						ID:                  teamMemberId,
						SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamMemberId.String()),
						Name:                fmt.Sprintf("%s Member", teamTemplate.Name),
						Description:         "Team Member Role",
						PermissionCodeNames: k.BASIC_PERMISSIONS(),
						CreatedBy:           user.ID.String(),
						UpdatedAt:           time.Now(),
						CreatedAt:           time.Now(),
						IsDefault:           false,
						OwnerType:           OwnerType(TEAMS),
						OwnerID:             teamID,
					},
				}

				// Initialize Users and Roles
				team.Users = []User{}
				team.Roles = []Role{}

				// MARK: Save Dept Team
				// Save team in database
				if err := o.BranchStore.DB.Create(&team).Error; err != nil {
					return nil, status.Errorf(500, "failed to create team: %v", err)
				}

				// Create roles for team
				for i := range teamRoles {
					if _, err := o.RoleStore.SaveRole(&teamRoles[i]); err != nil {
						return nil, status.Errorf(500, "failed to create team role: %v", err)
					}
				}

				// Add the team to the department's Teams collection
				if err := o.BranchStore.DB.Model(&dept).Association("Teams").Append(&team); err != nil {
					return nil, status.Errorf(500, "failed to associate team with department: %v", err)
				}

				// Create user-team association
				if err := o.BranchStore.DB.Model(&team).Association("Users").Append(user); err != nil {
					return nil, status.Errorf(500, "failed to add user to team: %v", err)
				}

				// Now add roles to the team and user separately
				for i := range teamRoles {
					// Add role to team
					if err := o.BranchStore.DB.Model(&team).Association("Roles").Append(&teamRoles[i]); err != nil {
						return nil, status.Errorf(500, "failed to add role to team: %v", err)
					}

					// Add role to user
					if err := o.BranchStore.DB.Model(user).Association("Roles").Append(&teamRoles[i]); err != nil {
						return nil, status.Errorf(500, "failed to add role to user: %v", err)
					}
				}
			}

			// MARK: Save Dept
			// Add department roles to department and user
			for i := range deptRoles {
				// Add role to department
				if err := o.BranchStore.DB.Model(&dept).Association("Roles").Append(&deptRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to department: %v", err)
				}

				// Add role to user
				if err := o.BranchStore.DB.Model(user).Association("Roles").Append(&deptRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to user: %v", err)
				}
			}

			// Create user-department association
			if err := o.BranchStore.DB.Model(&dept).Association("Users").Append(user); err != nil {
				return nil, status.Errorf(500, "failed to add user to department: %v", err)
			}
		}

		// MARK: Branches Teams
		for _, teamTemplate := range branchTemplate.Teams {
			teamID := uuid.New()
			team := Team{
				ID:           teamID,
				SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["teams"], teamID.String()),
				Name:         teamTemplate.Name,
				Description:  teamTemplate.Description,
				CreatedBy:    user.ID.String(),
				OwnerType:    OwnerType(BRANCHES),
				OwnerID:      branchID,
				Users:        []User{*user},
				Roles:        []Role{},
			}

			// Create team roles
			teamAdminId := uuid.New()
			teamMemberId := uuid.New()

			// MARK: Branch Team Roles
			teamRoles := []Role{
				{
					ID:                  teamAdminId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamAdminId.String()),
					Name:                fmt.Sprintf("%s Admin", teamTemplate.Name),
					Description:         "Team Admin Role",
					PermissionCodeNames: k.ADMIN_PERMISSIONS(),
					CreatedBy:           user.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           OwnerType(TEAMS),
					OwnerID:             teamID,
				},
				{
					ID:                  teamMemberId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamMemberId.String()),
					Name:                fmt.Sprintf("%s Member", teamTemplate.Name),
					Description:         "Team Member Role",
					PermissionCodeNames: k.BASIC_PERMISSIONS(),
					CreatedBy:           user.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           OwnerType(TEAMS),
					OwnerID:             teamID,
				},
			}

			// Initialize Users and Roles
			team.Users = []User{}
			team.Roles = []Role{}

			//MARK: Save Branch Team
			// Save team in database
			if err := o.BranchStore.DB.Create(&team).Error; err != nil {
				return nil, status.Errorf(500, "failed to create team: %v", err)
			}

			// Create roles for team
			for i := range teamRoles {
				if _, err := o.RoleStore.SaveRole(&teamRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to create team role: %v", err)
				}
			}

			// Add the team to the branch's Teams collection
			if err := o.BranchStore.DB.Model(&branch).Association("Teams").Append(&team); err != nil {
				return nil, status.Errorf(500, "failed to associate team with branch: %v", err)
			}

			// Create user-team association
			if err := o.BranchStore.DB.Model(&team).Association("Users").Append(user); err != nil {
				return nil, status.Errorf(500, "failed to add user to team: %v", err)
			}

			// Now add roles to the team and user separately
			for i := range teamRoles {
				// Add role to team
				if err := o.BranchStore.DB.Model(&team).Association("Roles").Append(&teamRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to team: %v", err)
				}

				// Add role to user
				if err := o.BranchStore.DB.Model(user).Association("Roles").Append(&teamRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to user: %v", err)
				}
			}
		}

		// MARK: Save Branch
		// Create user-branch association
		if err := o.BranchStore.DB.Model(&branch).Association("Users").Append(user); err != nil {
			return nil, status.Errorf(500, "failed to add user to branch: %v", err)
		}

		// Add roles to branch and user
		for i := range branchRoles {
			// Add role to branch
			if err := o.BranchStore.DB.Model(&branch).Association("Roles").Append(&branchRoles[i]); err != nil {
				return nil, status.Errorf(500, "failed to add role to branch: %v", err)
			}

			// Add role to user
			if err := o.BranchStore.DB.Model(user).Association("Roles").Append(&branchRoles[i]); err != nil {
				return nil, status.Errorf(500, "failed to add role to user: %v", err)
			}
		}

		// Add branch to organization
		if err := o.BranchStore.DB.Model(org).Association("Branches").Append(&branch); err != nil {
			return nil, status.Errorf(500, "failed to add branch to organization: %v", err)
		}

	}
	// Update organization
	if err := o.OrganisationStore.Update(org); err != nil {
		return nil, status.Errorf(500, "failed to update organization: %v", err)
	}

	// Get the updated organization
	updatedOrg, err := o.OrganisationStore.FindByID(&org.ID)
	if err != nil {
		return nil, status.Errorf(500, "failed to get updated organization: %v", err)
	}

	return &pb.CreateStructureWithTemplateResponse{
		Organisation: OrganisationToPbOrganisation(updatedOrg),
	}, nil
}
