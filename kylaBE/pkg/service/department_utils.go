package service

import (
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func DepartmentToPbDepartment(department *Department) *pb.Department {
	users := make([]*User, 0)
	for _, user := range department.Users {
		users = append(users, &user)
	}

	roles := make([]*Role, 0)
	for _, role := range department.Roles {
		roles = append(roles, &role)
	}

	return &pb.Department{
		Id:             department.ID.String(),
		DepartmentName: department.DepartmentName,
		SerialNumber:   department.SerialNumber,
		DepartmentBio:  department.DepartmentBio,
		CreatedAt:      department.CreatedAt.String(),
		Status:         department.Status,
		Users:          UsersToPbUsers(users),
		Roles:          RolesToPbRoles(roles),
		CreatedBy:      department.CreatedBy,
		UpdatedBy:      department.UpdatedBy,
		UpdatedAt:      department.UpdatedAt.String(),
		OwnerType:      pb.OwnerType(pb.OwnerType_value[string(department.OwnerType)]),
		OwnerId:        department.OwnerID.String(),
	}
}

func PbDepartmentToDepartment(department *pb.Department) *Department {
	id, err := uuid.Parse(department.GetId())
	if err != nil {
		id = uuid.New()
	}

	return &Department{
		ID:             id,
		SerialNumber:   utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["departments"], id.String()),
		DepartmentName: department.DepartmentName,
		DepartmentBio:  department.DepartmentBio,
		Status:         department.Status,
		CreatedBy:      department.CreatedBy,
		UpdatedBy:      department.UpdatedBy,
		Users:          PbUsersToUsers(department.Users),
		OwnerID:        uuid.MustParse(department.OwnerId),
		OwnerType:      OwnerType(department.OwnerType),
	}
}

func DepartmentsToPbDepartments(departments []*Department) []*pb.Department {
	var pbDepartments []*pb.Department
	for _, department := range departments {
		pbDepartments = append(pbDepartments, DepartmentToPbDepartment(department))
	}
	return pbDepartments
}

func PbDepartmentsToDepartments(departments []*pb.Department) []*Department {
	var pbDepartments []*Department
	for _, department := range departments {
		pbDepartments = append(pbDepartments, PbDepartmentToDepartment(department))
	}
	return pbDepartments
}
