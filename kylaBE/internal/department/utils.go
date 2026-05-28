package department

import (
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func DepartmentToPbDepartment(dept *Department) *pb.Department {
	roles := make([]*rbac.Role, 0)
	for _, role := range dept.Roles {
		r := role
		roles = append(roles, &r)
	}

	return &pb.Department{
		Id:             dept.ID.String(),
		DepartmentName: dept.DepartmentName,
		SerialNumber:   dept.SerialNumber,
		DepartmentBio:  dept.DepartmentBio,
		CreatedAt:      dept.CreatedAt.String(),
		Status:         dept.Status,
		// Users are populated separately (e.g. via DB preload) since the model
		// stores only UserRef (ID-only) to avoid import cycles.
		Users:     []*pb.User{},
		Roles:     rbac.RolesToPbRoles(roles),
		CreatedBy: dept.CreatedBy,
		UpdatedBy: dept.UpdatedBy,
		UpdatedAt: dept.UpdatedAt.String(),
		OwnerType: pb.OwnerType(pb.OwnerType_value[string(dept.OwnerType)]),
		OwnerId:   dept.OwnerID.String(),
	}
}

func PbDepartmentToDepartment(dept *pb.Department) *Department {
	id, err := uuid.Parse(dept.GetId())
	if err != nil {
		id = uuid.New()
	}

	// Convert pb.User IDs to UserRef entries.
	users := make([]UserRef, 0, len(dept.GetUsers()))
	for _, u := range dept.GetUsers() {
		uid, parseErr := uuid.Parse(u.GetId())
		if parseErr == nil && uid != uuid.Nil {
			users = append(users, UserRef{ID: uid})
		}
	}

	return &Department{
		ID:             id,
		SerialNumber:   utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["departments"], id.String()),
		DepartmentName: dept.DepartmentName,
		DepartmentBio:  dept.DepartmentBio,
		Status:         dept.Status,
		CreatedBy:      dept.CreatedBy,
		UpdatedBy:      dept.UpdatedBy,
		Users:          users,
		OwnerID:        uuid.MustParse(dept.OwnerId),
		OwnerType:      authctx.OwnerType(pb.OwnerType_name[int32(dept.OwnerType)]),
	}
}

func DepartmentsToPbDepartments(departments []*Department) []*pb.Department {
	var pbDepartments []*pb.Department
	for _, dept := range departments {
		pbDepartments = append(pbDepartments, DepartmentToPbDepartment(dept))
	}
	return pbDepartments
}

func PbDepartmentsToDepartments(departments []*pb.Department) []*Department {
	var result []*Department
	for _, dept := range departments {
		result = append(result, PbDepartmentToDepartment(dept))
	}
	return result
}
