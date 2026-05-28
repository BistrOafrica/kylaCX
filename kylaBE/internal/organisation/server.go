package organisation

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyla-be/internal/agentops"
	"kyla-be/internal/authctx"
	casbinsvc "kyla-be/internal/casbin"
	"kyla-be/internal/rbac"
	"kyla-be/internal/user"
	"kyla-be/internal/workspace"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// AuthGateway abstracts auth-metadata resolution needed by the Org server.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// Org implements the OrganisationService gRPC server.
type Org struct {
	pb.UnimplementedOrganisationServiceServer
	OrganisationStore *OrganisationStore
	AuthGateway       AuthGateway
	// DB is used directly for cross-domain raw GORM operations (branch/team/dept creation).
	DB             *gorm.DB
	RoleStore      *rbac.RbacStore
	UserStore      *user.UserStore
	AgentStore     *agentops.StatusStore
	EmailService   *utils.ResendService
	WorkspaceStore *workspace.WorkspaceStore
	CasbinEnforcer *casbinsvc.Enforcer // may be nil — seeds Casbin on org creation
}

// NewOrganisationServer constructs a new Org server.
func NewOrganisationServer(
	orgStore *OrganisationStore,
	authGateway AuthGateway,
	db *gorm.DB,
	roleStore *rbac.RbacStore,
	userStore *user.UserStore,
	agentStore *agentops.StatusStore,
	emailService *utils.ResendService,
	workspaceStore *workspace.WorkspaceStore,
	casbinEnforcer *casbinsvc.Enforcer,
) *Org {
	return &Org{
		OrganisationStore: orgStore,
		AuthGateway:       authGateway,
		DB:                db,
		RoleStore:         roleStore,
		UserStore:         userStore,
		AgentStore:        agentStore,
		EmailService:      emailService,
		WorkspaceStore:    workspaceStore,
		CasbinEnforcer:    casbinEnforcer,
	}
}

// CreateOrganisation handles the CreateOrganisation gRPC call.
func (o *Org) CreateOrganisation(ctx context.Context, request *pb.CreateOrganisationRequest) (*pb.CreateOrganisationResponse, error) {
	log.Println("Create Organisation")

	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	log.Println("Context Data: ", contextData)
	if authErr != nil || (contextData.RequestAuth != k.NewConsts().TRUE) {
		return nil, status.Errorf(403, "forbidden: you don't have permission to create an organisation%s, %s, %s", authErr, contextData.Authorization, k.NewConsts().TRUE)
	}

	u, er := o.UserStore.FindByID(&contextData.UserID)
	if er != nil {
		return nil, status.Error(500, "error while fetching user")
	}

	org := PbOrganisationToOrganisation(request.GetOrganisation())
	org.ID = uuid.New()
	org.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["organisations"], org.ID.String())
	org.Users = []user.User{*u}
	org.ADD_BASIC_ROLE()

	newOrg, err := o.OrganisationStore.Save(org)
	if err != nil {
		return nil, status.Error(500, "error while saving organisation")
	}

	// Auto-create a default "Support" workspace for the new organisation.
	var createdWS *workspace.Workspace
	if o.WorkspaceStore != nil {
		defaultWS := &workspace.Workspace{
			OrgID:          newOrg.ID.String(),
			Name:           "Support",
			Slug:           "support",
			Description:    "Default support workspace",
			DomainTemplate: workspace.DomainTemplateSupport,
			Status:         workspace.WorkspaceStatusActive,
		}
		ws, wsErr := o.WorkspaceStore.Create(defaultWS)
		if wsErr != nil {
			log.Printf("[org] warning: failed to create default workspace for org %s: %v", newOrg.ID, wsErr)
		} else {
			createdWS = ws
			// Add the creating user as owner of the default workspace.
			_, memErr := o.WorkspaceStore.AddMember(&workspace.WorkspaceMember{
				WorkspaceID: ws.ID,
				UserID:      u.ID.String(),
				Role:        workspace.MemberRoleOwner,
				JoinedAt:    time.Now(),
			})
			if memErr != nil {
				log.Printf("[org] warning: failed to add owner to workspace %s: %v", ws.ID, memErr)
			}
			log.Printf("[org] default workspace %s created for org %s", ws.ID, newOrg.ID)
		}
	}

	// Seed Casbin RBAC policies for the new organisation and workspace.
	if o.CasbinEnforcer != nil {
		if cErr := casbinsvc.SeedOrgAdmin(o.CasbinEnforcer, newOrg.ID.String(), u.ID.String()); cErr != nil {
			log.Printf("[org] casbin org-admin seed warning for org %s: %v", newOrg.ID, cErr)
		}
		if createdWS != nil {
			if cErr := casbinsvc.SeedWorkspaceOwner(o.CasbinEnforcer, createdWS.ID, newOrg.ID.String(), u.ID.String(), string(workspace.DomainTemplateSupport)); cErr != nil {
				log.Printf("[org] casbin ws-owner seed warning for ws %s: %v", createdWS.ID, cErr)
			}
		}
	}

	return &pb.CreateOrganisationResponse{
		Organisation: OrganisationToPbOrganisation(newOrg),
	}, nil
}

// ReadOrganisation handles the ReadOrganisation gRPC call.
func (o *Org) ReadOrganisation(ctx context.Context, request *pb.ReadOrganisationRequest) (*pb.ReadOrganisationResponse, error) {
	log.Println("Read Organisation")

	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	log.Println("Context Data: ", contextData)
	if authErr != nil || (contextData.RequestAuth != k.NewConsts().TRUE) {
		return nil, status.Errorf(403, "forbidden: you don't have permission to read an organisation%s, %s, %s", authErr, contextData.Authorization, k.NewConsts().TRUE)
	}

	var organisation *pb.Organisation
	if org, err := o.OrganisationStore.FindByID(&contextData.OrganisationID); err != nil {
		return nil, status.Error(500, "error while fetching organisation")
	} else {
		organisation = OrganisationToPbOrganisation(org)
	}

	return &pb.ReadOrganisationResponse{
		Organisation: organisation,
	}, nil
}

// UpdateOrganisation handles the UpdateOrganisation gRPC call.
func (o *Org) UpdateOrganisation(ctx context.Context, request *pb.UpdateOrganisationRequest) (*pb.UpdateOrganisationResponse, error) {
	log.Println("Update Organisation")

	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to update an organisation")
	}

	org := PbOrganisationToOrganisation(request.GetOrganisation())

	if err := utils.ValidateRequiredFields(k.NewConsts().OrganisationRequiredFields, org); err != nil {
		return nil, err
	}

	if err := o.OrganisationStore.Update(org); err != nil {
		return nil, status.Error(500, "error while updating organisation")
	}

	return &pb.UpdateOrganisationResponse{
		Organisation: OrganisationToPbOrganisation(org),
	}, nil
}

// ReadOrganisations handles the ReadOrganisations gRPC call.
func (o *Org) ReadOrganisations(ctx context.Context, request *pb.ReadOrganisationsRequest) (*pb.ReadOrganisationsResponse, error) {
	log.Println("Read Organisations")

	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to read organisations")
	}

	orgs, err := o.OrganisationStore.FindAll()
	if err != nil {
		return nil, status.Error(500, "internal server error")
	}

	return &pb.ReadOrganisationsResponse{
		Organisations: OrganisationsToPbOrganisations(orgs),
	}, nil
}

// DeleteOrganisation handles the DeleteOrganisation gRPC call.
func (o *Org) DeleteOrganisation(ctx context.Context, request *pb.DeleteOrganisationRequest) (*pb.DeleteOrganisationResponse, error) {
	log.Println("Delete Organisation")

	contextData, authErr := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to delete an organisation")
	}

	id, err := uuid.Parse(request.GetId())
	if err != nil {
		return nil, status.Error(400, "invalid organisation ID")
	}

	existingOrg, err := o.OrganisationStore.FindByID(&id)
	if err != nil {
		return nil, status.Error(500, "error while fetching organisation")
	}

	if err := o.OrganisationStore.Delete(existingOrg); err != nil {
		return nil, status.Error(500, "error while deleting organisation")
	}

	return &pb.DeleteOrganisationResponse{
		Success: true,
	}, nil
}

// ReadMe returns the authenticated user and their organisation.
func (o *Org) ReadMe(ctx context.Context, request *pb.ReadMeRequest) (*pb.ReadMeResponse, error) {
	log.Println("Read User")

	contextData, err := o.AuthGateway.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
	}

	if contextData.UserID == uuid.Nil {
		return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
	}

	u, err := o.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching user: %v", err)
	}
	org, err := o.OrganisationStore.FindByID(&contextData.OrganisationID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching organisation: %v", err)
	}

	return &pb.ReadMeResponse{
		User:         user.UserToPbUser(u),
		Organisation: OrganisationToPbOrganisation(org),
	}, nil
}
