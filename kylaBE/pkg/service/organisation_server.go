package service

import (
	"context"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type Org struct {
	pb.UnimplementedOrganisationServiceServer
	OrganisationStore *OrganisationStore
	AuthStore         *AuthStore
	BranchStore       *BranchStore
	RoleStore         *RbacStore
	UserStore         *UserStore
	AgentStore        *StatusStore
	EmailService      *utils.ResendService
}

func NewOrganisationServer(
	orgStore *OrganisationStore,
	authStore *AuthStore,
	branchStore *BranchStore,
	roleStore *RbacStore,
	userStore *UserStore,
	agentStore *StatusStore,
	emailService *utils.ResendService,
) *Org {
	return &Org{
		OrganisationStore: orgStore,
		AuthStore:         authStore,
		BranchStore:       branchStore,
		RoleStore:         roleStore,
		UserStore:         userStore,
		AgentStore:        agentStore,
		EmailService:      emailService,
	}
}

func (o *Org) CreateOrganisation(ctx context.Context, request *pb.CreateOrganisationRequest) (*pb.CreateOrganisationResponse, error) {
	log.Println("Create Organisation")

	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
	log.Println("Context Data: ", contextData)
	if authErr != nil || (contextData.RequestAuth != k.NewConsts().TRUE) {
		return nil, status.Errorf(403, "forbidden: you don't have permission to create an organisation%s, %s, %s", authErr, contextData.Authorization, k.NewConsts().TRUE)
	}

	user, er := o.UserStore.FindByID(&contextData.UserID)
	if er != nil {
		return nil, status.Error(500, "error while fetching user")
	}

	org := PbOrganisationToOrganisation(request.GetOrganisation())
	org.ID = uuid.New()
	org.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["organisations"], org.ID.String())
	org.Users = []User{*user}
	org.ADD_BASIC_ROLE()

	newOrg, err := o.OrganisationStore.Save(org)
	if err != nil {
		return nil, status.Error(500, "error while saving organisation")
	}

	return &pb.CreateOrganisationResponse{
		Organisation: OrganisationToPbOrganisation(newOrg),
	}, nil
}

func (o *Org) ReadOrganisation(ctx context.Context, request *pb.ReadOrganisationRequest) (*pb.ReadOrganisationResponse, error) {
	log.Println("Read Organisation")

	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
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

func (o *Org) UpdateOrganisation(ctx context.Context, request *pb.UpdateOrganisationRequest) (*pb.UpdateOrganisationResponse, error) {
	log.Println("Update Organisation")

	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to update an organisation")
	}

	org := PbOrganisationToOrganisation(request.GetOrganisation())

	// Validate required fields
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

func (o *Org) ReadOrganisations(ctx context.Context, request *pb.ReadOrganisationsRequest) (*pb.ReadOrganisationsResponse, error) {
	log.Println("Read Organisations")

	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "forbidden: you don't have permission to read organisations")
	}

	// Query all organisations
	orgs, err := o.OrganisationStore.FindAll()
	if err != nil {
		return nil, status.Error(500, "internal server error")
	}

	return &pb.ReadOrganisationsResponse{
		Organisations: OrganisationsToPbOrganisations(orgs),
	}, nil
}

func (o *Org) DeleteOrganisation(ctx context.Context, request *pb.DeleteOrganisationRequest) (*pb.DeleteOrganisationResponse, error) {
	log.Println("Delete Organisation")

	contextData, authErr := o.AuthStore.GetServiceAuthMetadata(ctx)
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

	// Delete the organisation from the database
	err = o.OrganisationStore.Delete(existingOrg)
	if err != nil {
		return nil, status.Error(500, "error while deleting organisation")
	}

	return &pb.DeleteOrganisationResponse{
		Success: true,
	}, nil
}

func (o *Org) ReadMe(ctx context.Context, request *pb.ReadMeRequest) (*pb.ReadMeResponse, error) {
	log.Println("Read User")

	contextData, err := o.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
	}

	if contextData.UserID == uuid.Nil {
		return nil, status.Error(403, "Forbidden, You do not have access to get user resource")
	}

	user, err := o.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching user: %v", err)
	}
	org, err := o.OrganisationStore.FindByID(&contextData.OrganisationID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching organisation: %v", err)
	}

	return &pb.ReadMeResponse{
		User:         UserToPbUser(user),
		Organisation: OrganisationToPbOrganisation(org),
	}, nil
}
