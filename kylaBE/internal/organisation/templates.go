package organisation

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyla-be/internal/authctx"
	"kyla-be/internal/branch"
	"kyla-be/internal/department"
	"kyla-be/internal/rbac"
	"kyla-be/internal/team"
	"kyla-be/internal/user"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

// ReadTemplates returns all organisation templates, optionally filtered by industry.
func (o *Org) ReadTemplates(ctx context.Context, request *pb.ReadTemplatesRequest) (*pb.ReadTemplatesResponse, error) {
	log.Println("Read Templates")

	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to read templates")
	}

	result := &pb.ReadTemplatesResponse{
		IndustryTemplates: []*pb.IndustryTemplates{},
	}

	for industry, subIndustryMap := range k.OrganizationTemplates {
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
			pbTemplate := &pb.OrganizationTemplate{
				Industry:            template.Industry,
				Subindustry:         template.Subindustry,
				TemplateName:        template.TemplateName,
				TemplateDescription: template.TemplateDescription,
				Branches:            []*pb.BranchTemplate{},
			}
			for _, b := range template.Branches {
				pbBranch := &pb.BranchTemplate{
					Name:        b.Name,
					Description: b.Description,
					Departments: []*pb.DepartmentTemplate{},
					Teams:       []*pb.TeamTemplate{},
				}
				for _, dept := range b.Departments {
					pbDept := &pb.DepartmentTemplate{
						Name:        dept.Name,
						Description: dept.Description,
						Teams:       []*pb.TeamTemplate{},
					}
					for _, t := range dept.Teams {
						pbDept.Teams = append(pbDept.Teams, &pb.TeamTemplate{
							Name:        t.Name,
							Description: t.Description,
						})
					}
					pbBranch.Departments = append(pbBranch.Departments, pbDept)
				}
				for _, t := range b.Teams {
					pbBranch.Teams = append(pbBranch.Teams, &pb.TeamTemplate{
						Name:        t.Name,
						Description: t.Description,
					})
				}
				pbTemplate.Branches = append(pbTemplate.Branches, pbBranch)
			}
			industryTemplates.SubindustryTemplates[subIndustry] = pbTemplate
		}
		result.IndustryTemplates = append(result.IndustryTemplates, industryTemplates)
	}

	return result, nil
}

// CreateDefaultStructure creates an organisation structure using the default template.
func (o *Org) CreateDefaultStructure(ctx context.Context, request *pb.CreateDefaultStructureRequest) (*pb.CreateDefaultStructureResponse, error) {
	log.Println("Create Default Structure")
	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to create organization structure")
	}

	orgID, err := uuid.Parse(request.OrganisationId)
	if err != nil {
		return nil, status.Error(400, "invalid organization ID")
	}

	org, err := o.OrganisationStore.FindByID(&orgID)
	if err != nil {
		return nil, status.Error(404, "organization not found")
	}

	u, err := o.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Error(500, "error while fetching user")
	}

	defaultTemplate := k.OrganizationTemplates["Default"]["Generic"]
	if defaultTemplate.Industry == "" {
		return nil, status.Error(500, "default template not found")
	}

	result, err := o.createStructure(u, org, defaultTemplate)
	if err != nil {
		return nil, err
	}

	return &pb.CreateDefaultStructureResponse{
		Organisation: result.Organisation,
	}, nil
}

// CreateStructureWithTemplate creates an organisation structure using the specified template.
func (o *Org) CreateStructureWithTemplate(ctx context.Context, request *pb.CreateStructureWithTemplateRequest) (*pb.CreateStructureWithTemplateResponse, error) {
	log.Println("Create Structure With Template")
	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to create organization structure")
	}

	orgID, err := uuid.Parse(request.OrganisationId)
	if err != nil {
		return nil, status.Error(400, "invalid organization ID")
	}

	org, err := o.OrganisationStore.FindByID(&orgID)
	if err != nil {
		return nil, status.Error(404, "organization not found")
	}

	u, err := o.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Error(500, "error while fetching user")
	}

	industryTemplates, ok := k.OrganizationTemplates[request.Industry]
	if !ok {
		return nil, status.Errorf(400, "industry '%s' not found", request.Industry)
	}

	subindustry := request.Subindustry
	if subindustry == "" {
		for sub := range industryTemplates {
			subindustry = sub
			break
		}
	}

	template, ok := industryTemplates[subindustry]
	if !ok {
		return nil, status.Errorf(400, "subindustry '%s' not found for industry '%s'", subindustry, request.Industry)
	}

	return o.createStructure(u, org, template)
}

// createStructure is the common implementation for both template creation methods.
func (o *Org) createStructure(u *user.User, org *Organisation, template k.OrganizationTemplate) (*pb.CreateStructureWithTemplateResponse, error) {

	for _, branchTemplate := range template.Branches {
		branchID := uuid.New()
		b := branch.Branch{
			ID:           branchID,
			SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["branches"], branchID.String()),
			Name:         branchTemplate.Name,
			Description:  branchTemplate.Description,
			CreatedBy:    u.ID.String(),
			Status:       "ACTIVE",
			ParentID:     uuid.Nil,
			IsDefault:    false,
			OwnerType:    authctx.ORGANISATIONS,
			OwnerID:      org.ID,
		}

		adminId := uuid.New()
		supervisorId := uuid.New()
		agentId := uuid.New()

		branchRoles := []rbac.Role{
			{
				ID:                  adminId,
				SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], adminId.String()),
				Name:                fmt.Sprintf("%s Admin", branchTemplate.Name),
				Description:         "Branch Admin Role",
				PermissionCodeNames: k.ADMIN_PERMISSIONS(),
				CreatedBy:           u.ID.String(),
				UpdatedAt:           time.Now(),
				CreatedAt:           time.Now(),
				IsDefault:           false,
				OwnerType:           authctx.BRANCHES,
				OwnerID:             branchID,
			},
			{
				ID:                  supervisorId,
				SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], supervisorId.String()),
				Name:                fmt.Sprintf("%s Supervisor", branchTemplate.Name),
				Description:         "Branch Supervisor Role",
				PermissionCodeNames: k.SUPERVISOR_PERMISSIONS(),
				CreatedBy:           u.ID.String(),
				UpdatedAt:           time.Now(),
				CreatedAt:           time.Now(),
				IsDefault:           false,
				OwnerType:           authctx.BRANCHES,
				OwnerID:             branchID,
			},
			{
				ID:                  agentId,
				SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], agentId.String()),
				Name:                fmt.Sprintf("%s Agent", branchTemplate.Name),
				Description:         "Branch Agent Role",
				PermissionCodeNames: k.AGENT_PERMISSIONS(),
				CreatedBy:           u.ID.String(),
				UpdatedAt:           time.Now(),
				CreatedAt:           time.Now(),
				IsDefault:           false,
				OwnerType:           authctx.BRANCHES,
				OwnerID:             branchID,
			},
		}

		if err := o.DB.Create(&b).Error; err != nil {
			return nil, status.Errorf(500, "failed to create branch: %v", err)
		}

		for i := range branchRoles {
			if _, err := o.RoleStore.SaveRole(&branchRoles[i]); err != nil {
				return nil, status.Errorf(500, "failed to create branch role: %v", err)
			}
		}

		// Branch departments
		for _, deptTemplate := range branchTemplate.Departments {
			deptID := uuid.New()
			dept := department.Department{
				ID:             deptID,
				SerialNumber:   utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["departments"], deptID.String()),
				DepartmentName: deptTemplate.Name,
				DepartmentBio:  deptTemplate.Description,
				CreatedBy:      u.ID.String(),
				Status:         "ACTIVE",
				OwnerType:      authctx.BRANCHES,
				OwnerID:        branchID,
				Roles:          []rbac.Role{},
				Teams:          []team.Team{},
			}

			deptAdminId := uuid.New()
			deptMemberId := uuid.New()

			deptRoles := []rbac.Role{
				{
					ID:                  deptAdminId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], deptAdminId.String()),
					Name:                fmt.Sprintf("%s Admin", deptTemplate.Name),
					Description:         "Department Admin Role",
					PermissionCodeNames: k.ADMIN_PERMISSIONS(),
					CreatedBy:           u.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           authctx.DEPARTMENTS,
					OwnerID:             deptID,
				},
				{
					ID:                  deptMemberId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], deptMemberId.String()),
					Name:                fmt.Sprintf("%s Member", deptTemplate.Name),
					Description:         "Department Member Role",
					PermissionCodeNames: k.BASIC_PERMISSIONS(),
					CreatedBy:           u.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           authctx.DEPARTMENTS,
					OwnerID:             deptID,
				},
			}

			if err := o.DB.Create(&dept).Error; err != nil {
				return nil, status.Errorf(500, "failed to create department: %v", err)
			}

			for i := range deptRoles {
				if _, err := o.RoleStore.SaveRole(&deptRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to create department role: %v", err)
				}
			}

			// Department teams
			for _, teamTemplate := range deptTemplate.Teams {
				teamID := uuid.New()
				t := team.Team{
					ID:           teamID,
					SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["teams"], teamID.String()),
					Name:         teamTemplate.Name,
					Description:  teamTemplate.Description,
					CreatedBy:    u.ID.String(),
					OwnerType:    authctx.DEPARTMENTS,
					OwnerID:      deptID,
					Roles:        []rbac.Role{},
				}

				teamAdminId := uuid.New()
				teamMemberId := uuid.New()

				teamRoles := []rbac.Role{
					{
						ID:                  teamAdminId,
						SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamAdminId.String()),
						Name:                fmt.Sprintf("%s Admin", teamTemplate.Name),
						Description:         "Team Admin Role",
						PermissionCodeNames: k.ADMIN_PERMISSIONS(),
						CreatedBy:           u.ID.String(),
						UpdatedAt:           time.Now(),
						CreatedAt:           time.Now(),
						IsDefault:           false,
						OwnerType:           authctx.TEAMS,
						OwnerID:             teamID,
					},
					{
						ID:                  teamMemberId,
						SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamMemberId.String()),
						Name:                fmt.Sprintf("%s Member", teamTemplate.Name),
						Description:         "Team Member Role",
						PermissionCodeNames: k.BASIC_PERMISSIONS(),
						CreatedBy:           u.ID.String(),
						UpdatedAt:           time.Now(),
						CreatedAt:           time.Now(),
						IsDefault:           false,
						OwnerType:           authctx.TEAMS,
						OwnerID:             teamID,
					},
				}

				if err := o.DB.Create(&t).Error; err != nil {
					return nil, status.Errorf(500, "failed to create team: %v", err)
				}

				for i := range teamRoles {
					if _, err := o.RoleStore.SaveRole(&teamRoles[i]); err != nil {
						return nil, status.Errorf(500, "failed to create team role: %v", err)
					}
				}

				if err := o.DB.Model(&dept).Association("Teams").Append(&t); err != nil {
					return nil, status.Errorf(500, "failed to associate team with department: %v", err)
				}
				if err := o.DB.Model(&t).Association("Users").Append(u); err != nil {
					return nil, status.Errorf(500, "failed to add user to team: %v", err)
				}
				for i := range teamRoles {
					if err := o.DB.Model(&t).Association("Roles").Append(&teamRoles[i]); err != nil {
						return nil, status.Errorf(500, "failed to add role to team: %v", err)
					}
					if err := o.DB.Model(u).Association("Roles").Append(&teamRoles[i]); err != nil {
						return nil, status.Errorf(500, "failed to add role to user: %v", err)
					}
				}
			}

			for i := range deptRoles {
				if err := o.DB.Model(&dept).Association("Roles").Append(&deptRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to department: %v", err)
				}
				if err := o.DB.Model(u).Association("Roles").Append(&deptRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to user: %v", err)
				}
			}
			if err := o.DB.Model(&dept).Association("Users").Append(u); err != nil {
				return nil, status.Errorf(500, "failed to add user to department: %v", err)
			}
		}

		// Branch teams
		for _, teamTemplate := range branchTemplate.Teams {
			teamID := uuid.New()
			t := team.Team{
				ID:           teamID,
				SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["teams"], teamID.String()),
				Name:         teamTemplate.Name,
				Description:  teamTemplate.Description,
				CreatedBy:    u.ID.String(),
				OwnerType:    authctx.BRANCHES,
				OwnerID:      branchID,
				Roles:        []rbac.Role{},
			}

			teamAdminId := uuid.New()
			teamMemberId := uuid.New()

			teamRoles := []rbac.Role{
				{
					ID:                  teamAdminId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamAdminId.String()),
					Name:                fmt.Sprintf("%s Admin", teamTemplate.Name),
					Description:         "Team Admin Role",
					PermissionCodeNames: k.ADMIN_PERMISSIONS(),
					CreatedBy:           u.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           authctx.TEAMS,
					OwnerID:             teamID,
				},
				{
					ID:                  teamMemberId,
					SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], teamMemberId.String()),
					Name:                fmt.Sprintf("%s Member", teamTemplate.Name),
					Description:         "Team Member Role",
					PermissionCodeNames: k.BASIC_PERMISSIONS(),
					CreatedBy:           u.ID.String(),
					UpdatedAt:           time.Now(),
					CreatedAt:           time.Now(),
					IsDefault:           false,
					OwnerType:           authctx.TEAMS,
					OwnerID:             teamID,
				},
			}

			if err := o.DB.Create(&t).Error; err != nil {
				return nil, status.Errorf(500, "failed to create team: %v", err)
			}

			for i := range teamRoles {
				if _, err := o.RoleStore.SaveRole(&teamRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to create team role: %v", err)
				}
			}

			if err := o.DB.Model(&b).Association("Teams").Append(&t); err != nil {
				return nil, status.Errorf(500, "failed to associate team with branch: %v", err)
			}
			if err := o.DB.Model(&t).Association("Users").Append(u); err != nil {
				return nil, status.Errorf(500, "failed to add user to team: %v", err)
			}
			for i := range teamRoles {
				if err := o.DB.Model(&t).Association("Roles").Append(&teamRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to team: %v", err)
				}
				if err := o.DB.Model(u).Association("Roles").Append(&teamRoles[i]); err != nil {
					return nil, status.Errorf(500, "failed to add role to user: %v", err)
				}
			}
		}

		// Finalise branch: associate user and roles
		if err := o.DB.Model(&b).Association("Users").Append(u); err != nil {
			return nil, status.Errorf(500, "failed to add user to branch: %v", err)
		}
		for i := range branchRoles {
			if err := o.DB.Model(&b).Association("Roles").Append(&branchRoles[i]); err != nil {
				return nil, status.Errorf(500, "failed to add role to branch: %v", err)
			}
			if err := o.DB.Model(u).Association("Roles").Append(&branchRoles[i]); err != nil {
				return nil, status.Errorf(500, "failed to add role to user: %v", err)
			}
		}

		if err := o.DB.Model(org).Association("Branches").Append(&b); err != nil {
			return nil, status.Errorf(500, "failed to add branch to organization: %v", err)
		}
	}

	if err := o.OrganisationStore.Update(org); err != nil {
		return nil, status.Errorf(500, "failed to update organization: %v", err)
	}

	updatedOrg, err := o.OrganisationStore.FindByID(&org.ID)
	if err != nil {
		return nil, status.Errorf(500, "failed to get updated organization: %v", err)
	}

	return &pb.CreateStructureWithTemplateResponse{
		Organisation: OrganisationToPbOrganisation(updatedOrg),
	}, nil
}
