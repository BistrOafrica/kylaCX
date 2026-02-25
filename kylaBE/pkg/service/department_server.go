package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type DepartmentServer struct {
	pb.UnimplementedDepartmentServiceServer
	DepartmentStore *DepartmentStore
	AuthStore       *AuthStore
	BranchStore     *BranchStore
	UserStore       *UserStore
}

func NewDepartmentServer(DepartmentStore *DepartmentStore, AuthStore *AuthStore, branchStore *BranchStore, UserStore *UserStore) *DepartmentServer {
	return &DepartmentServer{
		DepartmentStore: DepartmentStore,
		AuthStore:       AuthStore,
		BranchStore:     branchStore,
		UserStore:       UserStore,
	}
}

func (d *DepartmentServer) CreateDepartment(ctx context.Context, request *pb.CreateDepartmentRequest) (*pb.CreateDepartmentResponse, error) {
	log.Println("Create Department")
	scope := PbScopeToOpScope(request.GetScope())
	contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go d.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to create department resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)

	}

	data := PbDepartmentToDepartment(request.GetDepartment())
	data.ID = uuid.New()
	data.CreatedBy = contextData.UserID.String()
	data.UpdatedBy = contextData.UserID.String()
	data.Status = k.GENERAL_STATUSES()["ACTIVE"]
	data.OwnerID = uuid.MustParse(scope.ID)
	data.OwnerType = OwnerType(scope.Owner)

	data.ADD_CREATOR_TO_DEPARTMENT(d.UserStore)
	data.CREATE_DEPARTMENT_ROLES(d.AuthStore.RbacStore)

	department, saveErr := d.DepartmentStore.CreateDepartment(data)
	if saveErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to create department")
	}

	pbDepartment := &pb.Department{}
	pbDepartmentChan := make(chan *pb.Department)
	errChan = make(chan error)

	go d.getPbDepartment(department, pbDepartmentChan, errChan)

	select {
	case pbDepartmentData := <-pbDepartmentChan:
		pbDepartment = pbDepartmentData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}
	return &pb.CreateDepartmentResponse{Department: pbDepartment}, nil
}

func (d *DepartmentServer) ReadDepartment(ctx context.Context, request *pb.ReadDepartmentRequest) (*pb.ReadDepartmentResponse, error) {
	log.Println("Read Department")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go d.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to read department resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	department, readErr := d.DepartmentStore.ReadDepartment(request.GetId())
	if readErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to read department")
	}

	pbDepartmentChan := make(chan *pb.Department)
	errChan = make(chan error)
	go d.getPbDepartment(department, pbDepartmentChan, errChan)

	pbDepartment := &pb.Department{}
	select {
	case pbDepartmentData := <-pbDepartmentChan:
		pbDepartment = pbDepartmentData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	return &pb.ReadDepartmentResponse{Department: pbDepartment}, nil
}

func (d *DepartmentServer) ReadDepartments(ctx context.Context, request *pb.ReadDepartmentsRequest) (*pb.ReadDepartmentsResponse, error) {
	log.Println("Read Departments")
	scope := PbScopeToOpScope(request.GetScope())
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go d.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to read departments resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	departments, readErr := d.DepartmentStore.ReadDepartments(scope)
	if readErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to read all departments")
	}

	pbDepartments := make([]*pb.Department, len(departments))

	for i, department := range departments {
		pbDepartmentChan := make(chan *pb.Department)
		errChan := make(chan error)
		go d.getPbDepartment(department, pbDepartmentChan, errChan)

		select {
		case pbDepartmentData := <-pbDepartmentChan:
			pbDepartments[i] = pbDepartmentData
		case err := <-errChan:
			return nil, status.Errorf(500, "error while fetching user %v", err)
		}
	}
	return &pb.ReadDepartmentsResponse{Departments: pbDepartments}, nil
}

func (d *DepartmentServer) DeleteDepartment(ctx context.Context, request *pb.DeleteDepartmentRequest) (*pb.DeleteDepartmentResponse, error) {
	log.Println("Read Departments")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go d.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to read departments resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	if deleteErr := d.DepartmentStore.DeleteDepartment(request.GetId()); deleteErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to delete department")
	}

	return &pb.DeleteDepartmentResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Department deleted successfully",
		},
	}, nil
}

func (d *DepartmentServer) UpdateDepartment(ctx context.Context, request *pb.UpdateDepartmentRequest) (*pb.UpdateDepartmentResponse, error) {
	log.Println("Read Departments")

	contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go d.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to Update department resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	department := PbDepartmentToDepartment(request.GetDepartment())
	department.UpdatedBy = contextData.UserID.String()

	updatedDepartment, updateErr := d.DepartmentStore.UpdateDepartment(department)
	if updateErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to update department")
	}

	pbDepartmentChan := make(chan *pb.Department)
	errChan = make(chan error)

	pbDepartment := &pb.Department{}
	go d.getPbDepartment(updatedDepartment, pbDepartmentChan, errChan)

	select {
	case pbDepartmentData := <-pbDepartmentChan:
		pbDepartment = pbDepartmentData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	return &pb.UpdateDepartmentResponse{Department: pbDepartment}, nil
}

func (d *DepartmentServer) getPbDepartment(department *Department, pbDepartmentChan chan *pb.Department, errChan chan error) {
	// Read department users
	usersChan := make(chan []User)
	select {
	case users := <-usersChan:
		pbDepartment := DepartmentToPbDepartment(department)
		for _, user := range users {
			pbDepartment.Users = append(pbDepartment.Users, UserToPbUser(&user))
		}
		pbDepartmentChan <- pbDepartment
	case err := <-errChan:
		errChan <- err
	}
}

func (d *DepartmentServer) ReadBranchDepartments(ctx context.Context, request *pb.ReadBranchDepartmentsRequest) (*pb.ReadBranchDepartmentsResponse, error) {
	log.Println("Read Departments by Branch")

	contextData := &RequestMetadata{}
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go d.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to read departments resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	departments, readErr := d.DepartmentStore.ReadDepartmentsByBranch(request.GetBranchId(), contextData.OrganisationID.String())
	if readErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to read all departments")
	}

	pbDepartments := make([]*pb.Department, len(departments))

	for i, department := range departments {
		pbDepartmentChan := make(chan *pb.Department)
		errChan := make(chan error)
		go d.getPbDepartment(department, pbDepartmentChan, errChan)

		select {
		case pbDepartmentData := <-pbDepartmentChan:
			pbDepartments[i] = pbDepartmentData
		case err := <-errChan:
			return nil, status.Errorf(500, "error while fetching user %v", err)
		}
	}
	return &pb.ReadBranchDepartmentsResponse{Departments: pbDepartments}, nil
}
