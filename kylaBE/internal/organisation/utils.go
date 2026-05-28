package organisation

import (
	"kyla-be/internal/branch"
	"kyla-be/internal/rbac"
	"kyla-be/internal/user"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// BranchToPbBranch converts a branch.Branch to its protobuf representation.
// User, department and team ID lists are populated from the relationships present
// in memory; full hydration requires preloaded associations.
func BranchToPbBranch(b *branch.Branch) *pb.Branch {
	// branch.Branch does not carry Roles/Teams/Departments slices (import cycle prevention).
	// Return empty slices; full hydration must be done via separate store queries.
	return &pb.Branch{
		Id:            b.ID.String(),
		Name:          b.Name,
		SerialNumber:  b.SerialNumber,
		IsDefault:     b.IsDefault,
		Description:   b.Description,
		Status:        b.Status,
		ParentId:      b.ParentID.String(),
		UserIds:       []string{},
		RoleIds:       []string{},
		Location:      b.Location,
		Address:       b.Address,
		DepartmentIds: []string{},
		TeamIds:       []string{},
		OwnerType:     pb.OwnerType(pb.OwnerType_value[string(b.OwnerType)]),
		OwnerId:       b.OwnerID.String(),
		CreatedBy:     b.CreatedBy,
		UpdatedBy:     b.UpdatedBy,
		CreatedAt:     b.CreatedAt.String(),
		UpdatedAt:     b.UpdatedAt.String(),
	}
}

// OrganisationToPbOrganisation converts an Organisation model to its protobuf form.
func OrganisationToPbOrganisation(org *Organisation) *pb.Organisation {
	pbRoles := make([]*pb.Role, 0, len(org.Roles))
	for i := range org.Roles {
		pbRoles = append(pbRoles, rbac.RoleToPbRole(&org.Roles[i]))
	}

	pbUsers := make([]*pb.User, 0, len(org.Users))
	for i := range org.Users {
		pbUsers = append(pbUsers, user.UserToPbUser(&org.Users[i]))
	}

	pbBranches := make([]*pb.Branch, 0, len(org.Branches))
	for i := range org.Branches {
		pbBranches = append(pbBranches, BranchToPbBranch(&org.Branches[i]))
	}

	return &pb.Organisation{
		Id:               org.ID.String(),
		OrganisationName: org.OrganisationName,
		SerialNumber:     org.SerialNumber,
		OrganisationBio:  org.OrganisationBio,
		Status:           org.Status,
		ReferralCode:     org.ReferralCode,
		ShortCode:        org.ShortCode,
		Email:            org.Email,
		Phone:            org.Phone,
		Size:             org.Size,
		Country:          org.Country,
		Industry:         org.Industry,
		SubIndustry:      org.SubIndustry,
		Roles:            pbRoles,
		Users:            pbUsers,
		Departments:      []*pb.Department{},
		Branches:         pbBranches,
		CreatedAt:        org.CreatedAt.String(),
	}
}

// PbOrganisationToOrganisation converts a protobuf Organisation to a model.
func PbOrganisationToOrganisation(pbOrg *pb.Organisation) *Organisation {
	id, err := uuid.Parse(pbOrg.Id)
	if err != nil {
		id = uuid.New()
	}
	if pbOrg.SerialNumber == "" {
		pbOrg.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["organisations"], id.String())
	}
	return &Organisation{
		ID:               id,
		OrganisationName: pbOrg.OrganisationName,
		SerialNumber:     pbOrg.SerialNumber,
		OrganisationBio:  pbOrg.OrganisationBio,
		Status:           pbOrg.Status,
		ReferralCode:     pbOrg.ReferralCode,
		ShortCode:        pbOrg.ShortCode,
		Email:            pbOrg.Email,
		Phone:            pbOrg.Phone,
		Size:             pbOrg.Size,
		Country:          pbOrg.Country,
		Industry:         pbOrg.Industry,
		SubIndustry:      pbOrg.SubIndustry,
	}
}

// OrganisationsToPbOrganisations converts a slice of Organisation pointers to protobuf.
func OrganisationsToPbOrganisations(orgs []*Organisation) []*pb.Organisation {
	result := make([]*pb.Organisation, 0, len(orgs))
	for _, org := range orgs {
		result = append(result, OrganisationToPbOrganisation(org))
	}
	return result
}

// PbOrganisationsToOrganisations converts a slice of protobuf Organisations to models.
func PbOrganisationsToOrganisations(orgs []*pb.Organisation) []*Organisation {
	result := make([]*Organisation, 0, len(orgs))
	for _, org := range orgs {
		result = append(result, PbOrganisationToOrganisation(org))
	}
	return result
}

// IsPbValidOwnerType returns true if ownerType is a defined proto enum value.
func IsPbValidOwnerType(ownerType pb.OwnerType) bool {
	enumDesc := pb.OwnerType(0).Descriptor()
	return enumDesc.Values().ByNumber(protoreflect.EnumNumber(ownerType)) != nil
}
