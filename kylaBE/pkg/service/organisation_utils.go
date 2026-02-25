package service

import (
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func OrganisationToPbOrganisation(org *Organisation) *pb.Organisation {
	roles := []*Role{}
	for _, role := range org.Roles {
		roles = append(roles, &role)
	}

	users := []*User{}
	for _, user := range org.Users {
		users = append(users, &user)
	}
	departments := []*Department{}
	for _, department := range org.Departments {
		departments = append(departments, &department)
	}

	branches := []*Branch{}
	for _, branch := range org.Branches {
		branches = append(branches, &branch)
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
		Roles:            RolesToPbRoles(roles),
		Users:            UsersToPbUsers(users),
		Departments:      DepartmentsToPbDepartments(departments),
		Branches:         BranchesToPbBranches(branches),
		CreatedAt:        org.CreatedAt.String(),
	}
}

func PbOrganisationToOrganisation(org *pb.Organisation) *Organisation {
	id, err := uuid.Parse(org.Id)
	if err != nil {
		id = uuid.New()
	}
	if org.SerialNumber == "" {
		org.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["organisations"], id.String())
	}
	return &Organisation{
		ID:               id,
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
	}
}

func OrganisationsToPbOrganisations(orgs []*Organisation) []*pb.Organisation {
	pbOrgs := []*pb.Organisation{}
	for _, org := range orgs {
		pbOrgs = append(pbOrgs, OrganisationToPbOrganisation(org))
	}
	return pbOrgs
}

func PbOrganisationsToOrganisations(orgs []*pb.Organisation) []*Organisation {
	Orgs := []*Organisation{}
	for _, org := range orgs {
		Orgs = append(Orgs, PbOrganisationToOrganisation(org))
	}
	return Orgs
}

// func ReadUserFromPbOrganisation(org *pb.Organisation) *User {
// 	return &User{
// 		FirstName: org.FirstName,
// 		LastName:  org.LastName,
// 		Username:  org.Username,
// 		Email:     org.Email,
// 		Phone:     org.Phone,
// 		Roles:     []Role{},
// 		Status:    k.USER_STATUSES()["NEW"],
// 		IsDefault: k.NewConsts().TRUE_BOOL,
// 		CreatedBy: "USERS",
// 	}
// }

func PbScopeToOpScope(scope *pb.Scope) *OpScope {
	return &OpScope{
		Owner: OwnerType(pb.OwnerType_name[int32(scope.OwnerType)]),
		ID:    scope.OwnerId,
	}
}

func IsPbValidOwnerType(ownerType pb.OwnerType) bool {
	// Get the enum descriptor for OwnerType
	enumDesc := pb.OwnerType(0).Descriptor()

	// Check if the value is defined in the enum
	return enumDesc.Values().ByNumber(protoreflect.EnumNumber(ownerType)) != nil
}
