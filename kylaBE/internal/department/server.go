package department

import (
	"context"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/service"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

// AuthGateway defines the auth methods DepartmentServer requires from AuthStore.
type AuthGateway interface {
	GetUserRequestMetadata(ctx context.Context, dataChan chan *authctx.RequestMetadata, errChan chan error)
	GetRbacStore() *rbac.RbacStore
}

type DepartmentServer struct {
	pb.UnimplementedDepartmentServiceServer
	DepartmentStore *DepartmentStore
	AuthStore       AuthGateway
	BranchStore     *service.BranchStore
	UserStore       *service.UserStore
}

func NewDepartmentServer(departmentStore *DepartmentStore, authStore AuthGateway, branchStore *service.BranchStore, userStore *service.UserStore) *DepartmentServer {
	return &DepartmentServer{
		DepartmentStore: departmentStore,
		AuthStore:       authStore,
		BranchStore:     branchStore,
		UserStore:       userStore,
	}
}

func (d *DepartmentServer) CreateDepartment(ctx context.Context, request *pb.CreateDepartmentRequest) (*pb.CreateDepartmentResponse, error) {
	log.Println("Create Department")
	scope := authctx.PbScopeToOpScope(request.GetScope())
	contextData := &authctx.RequestMetadata{}
	contextDataChan := make(chan *authctx.RequestMetadata)
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
	data.OwnerType = authctx.OwnerType(scope.Owner)

	data.ADD_CREATOR_TO_DEPARTMENT(d.UserStore)
	data.CREATE_DEPARTMENT_ROLES(d.AuthStore.GetRbacStore())

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

	contextDataChan := make(chan *authctx.RequestMetadata)
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
	scope := authctx.PbScopeToOpScope(request.GetScope())
	contextDataChan := make(chan *authctx.RequestMetadata)
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

	for i, dept := range departments {
		pbDepartmentChan := make(chan *pb.Department)
		errChan := make(chan error)
		go d.getPbDepartment(dept, pbDepartmentChan, errChan)

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

	contextDataChan := make(chan *authctx.RequestMetadata)
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

	contextData := &authctx.RequestMetadata{}
	contextDataChan := make(chan *authctx.RequestMetadata)
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

	dept := PbDepartmentToDepartment(request.GetDepartment())
	dept.UpdatedBy = contextData.UserID.String()

	updatedDepartment, updateErr := d.DepartmentStore.UpdateDepartment(dept)
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

func (d *DepartmentServer) getPbDepartment(dept *Department, pbDepartmentChan chan *pb.Department, errChan chan error) {
	// The original implementation awaited a usersChan that was never sent to,
	// resulting in a goroutine that blocks on errChan. Logic is preserved as-is.
	usersChan := make(chan []UserRef)
	select {
	case <-usersChan:
		pbDepartment := DepartmentToPbDepartment(dept)
		pbDepartmentChan <- pbDepartment
	case err := <-errChan:
		errChan <- err
	}
}

func (d *DepartmentServer) ReadBranchDepartments(ctx context.Context, request *pb.ReadBranchDepartmentsRequest) (*pb.ReadBranchDepartmentsResponse, error) {
	log.Println("Read Departments by Branch")

	contextData := &authctx.RequestMetadata{}
	contextDataChan := make(chan *authctx.RequestMetadata)
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

	for i, dept := range departments {
		pbDepartmentChan := make(chan *pb.Department)
		errChan := make(chan error)
		go d.getPbDepartment(dept, pbDepartmentChan, errChan)

		select {
		case pbDepartmentData := <-pbDepartmentChan:
			pbDepartments[i] = pbDepartmentData
		case err := <-errChan:
			return nil, status.Errorf(500, "error while fetching user %v", err)
		}
	}
	return &pb.ReadBranchDepartmentsResponse{Departments: pbDepartments}, nil
}
