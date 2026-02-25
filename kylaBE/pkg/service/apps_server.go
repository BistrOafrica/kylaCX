package service

import (
	"context"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/templates"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type AppServer struct {
	pb.UnimplementedAppServiceServer
	AuthStore    *AuthStore
	AppStore     *AppStore
	UserStore    *UserStore
	EmailService *utils.ResendService
}

func NewAppServer(
	AuthStore *AuthStore,
	AppStore *AppStore,
	UserStore *UserStore,
	EmailService *utils.ResendService,
) *AppServer {
	return &AppServer{
		AuthStore:    AuthStore,
		AppStore:     AppStore,
		UserStore:    UserStore,
		EmailService: EmailService,
	}
}

func (s *AppServer) CreateApp(ctx context.Context, req *pb.CreateAppRequest) (*pb.CreateAppResponse, error) {
	log.Println("Create App")
	scope := PbScopeToOpScope(req.GetScope())
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	var secret string
	{
	CREATE_SECRET:
		newSecret, err := utils.GENERATE_RANDOM_KEY(64)
		if err != nil {
			goto CREATE_SECRET
		}
		secret = newSecret
	}

	app := PbAppToApp(req.GetApp(), scope)
	app.ID = uuid.New()
	app.Status = k.GENERAL_STATUSES()["NEW"]
	app.CreatedBy = contextData.UserID.String()
	app.UpdatedBy = contextData.UserID.String()
	app.Token = uuid.New().String()
	app.Secret = utils.HASH_PASSWORD(secret)
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	app.OwnerType = scope.Owner
	app.OwnerId = scope.ID

	newApp, err := s.AppStore.SaveApp(app)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to create app %v", err)
	}
	newApp.Secret = secret
	newApp.Token = app.Token
	// Send Token and Secret to the user via Email
	creatorID := uuid.MustParse(app.CreatedBy)
	user, err := s.UserStore.FindByID(&creatorID)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to get user %v", err)
	}
	if emailErr := s.EmailService.SEND_APP_TOKEN_AND_SECRET_EMAIL(templates.AppSecretData{
		ClientEmail:  user.Email,
		Name:         user.FirstName,
		AppID:        newApp.ID.String(),
		AppName:      newApp.Name,
		Token:        newApp.Token,
		Secret:       secret,
		SupportEmail: fmt.Sprintln("support@kyla.com"),
		Year:         fmt.Sprintf("%d", time.Now().Year()),
	}); emailErr != nil {
		log.Printf("Error sending welcome email: %v", err)
	}
	return &pb.CreateAppResponse{
		App: AppToPbApp(newApp, scope),
	}, nil
}

func (s *AppServer) CreateAppWithTemplate(ctx context.Context, req *pb.CreateAppWithTemplateRequest) (*pb.CreateAppWithTemplateResponse, error) {
	log.Println("Create App with Template")
	scope := PbScopeToOpScope(req.GetScope())
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	app, err := s.AppStore.FindAppByID(req.GetTemplateAppId(), scope)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to get template app %v", err)
	}

	var secret string
	{
	CREATE_SECRET:
		newSecret, err := utils.GENERATE_RANDOM_KEY(64)
		if err != nil {
			goto CREATE_SECRET
		}
		secret = newSecret
	}

	app.ID = uuid.New()
	app.Status = k.GENERAL_STATUSES()["NEW"]
	app.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["apps"], app.ID.String())
	app.IsTemplate = false
	app.CreatedBy = contextData.UserID.String()
	app.UpdatedBy = contextData.UserID.String()
	app.Token = uuid.New().String()
	app.Secret = utils.HASH_PASSWORD(secret)
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	app.OwnerType = scope.Owner
	app.OwnerId = scope.ID

	newApp, err := s.AppStore.SaveApp(app)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to create app %v", err)
	}
	newApp.Secret = "We do not store the app secret, Please check your email and keep it safe and secure."
	newApp.Token = app.Token
	// Send Token and Secret to the user via Email
	user, err := s.UserStore.FindByID(&contextData.UserID)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to get user %v", err)
	}
	if emailErr := s.EmailService.SEND_APP_TOKEN_AND_SECRET_EMAIL(templates.AppSecretData{
		ClientEmail:  user.Email,
		Name:         user.FirstName,
		AppID:        newApp.ID.String(),
		AppName:      newApp.Name,
		Token:        newApp.Token,
		Secret:       secret,
		SupportEmail: fmt.Sprintln("support@kyla.com"),
		Year:         fmt.Sprintf("%d", time.Now().Year()),
	}); emailErr != nil {
		log.Printf("Error sending welcome email: %v", err)
	}

	return &pb.CreateAppWithTemplateResponse{
		App: AppToPbApp(newApp, scope),
	}, nil
}

func (s *AppServer) RegenerateAppKeyAndSecret(ctx context.Context, req *pb.UpdateAppRequest) (*pb.UpdateAppResponse, error) {
	log.Println("Regenerate App Key and secret")
	scope := PbScopeToOpScope(req.GetScope())
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}
	app := PbAppToApp(req.GetApp(), scope)
	app.UpdatedBy = contextData.UserID.String()

	var secret string
	{
	CREATE_SECRET:
		newSecret, err := utils.GENERATE_RANDOM_KEY(64)
		if err != nil {
			goto CREATE_SECRET
		}
		secret = newSecret
	}

	app.Token = uuid.New().String()
	app.Secret = string(utils.HASH_PASSWORD(secret))
	app.UpdatedAt = time.Now()

	newApp, err := s.AppStore.UpdateTokenAndSecret(app)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to regenerate app key and secret")
	}
	newApp.Secret = secret
	// Send Token and Secret to the user via Email
	creatorID := uuid.MustParse(app.CreatedBy)
	user, err := s.UserStore.FindByID(&creatorID)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to get user %v", err)
	}

	if emailErr := s.EmailService.SEND_APP_TOKEN_AND_SECRET_EMAIL(templates.AppSecretData{
		ClientEmail:  user.Email,
		Name:         user.FirstName,
		AppID:        newApp.ID.String(),
		AppName:      newApp.Name,
		Token:        newApp.Token,
		Secret:       secret,
		SupportEmail: fmt.Sprintln("support@kyla.com"),
		Year:         fmt.Sprintf("%d", time.Now().Year()),
	}); emailErr != nil {
		log.Printf("Error sending welcome email: %v", err)
	}

	return &pb.UpdateAppResponse{
		App: AppToPbApp(newApp, scope),
	}, nil
}

func (s *AppServer) ReadApp(ctx context.Context, req *pb.ReadAppRequest) (*pb.ReadAppResponse, error) {
	log.Println("Read App")

	scope := PbScopeToOpScope(req.GetScope())
	auth, _, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}
	app, err := s.AppStore.FindAppByID(req.GetId(), scope)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to get app")
	}
	app.Secret = ""

	return &pb.ReadAppResponse{
		App: AppToPbApp(app, scope),
	}, nil
}

func (s *AppServer) ReadApps(ctx context.Context, req *pb.ReadAppsRequest) (*pb.ReadAppsResponse, error) {
	log.Println("Read Apps")

	scope := PbScopeToOpScope(req.GetScope())
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	apps, err := s.AppStore.FindAllAppsByOrganisationID(contextData.OrganisationID)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to get apps")
	}
	// Hide Secret
	for _, app := range apps {
		app.Secret = ""
	}

	return &pb.ReadAppsResponse{
		Apps: AppsToPbApps(apps, scope),
	}, nil
}

func (s *AppServer) UpdateApp(ctx context.Context, req *pb.UpdateAppRequest) (*pb.UpdateAppResponse, error) {
	log.Println("Update App")

	scope := PbScopeToOpScope(req.GetScope())
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	app := PbAppToApp(req.GetApp(), scope)
	app.UpdatedBy = contextData.UserID.String()

	updatedApp, err := s.AppStore.UpdateApp(app)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to update app")
	}

	return &pb.UpdateAppResponse{
		App: AppToPbApp(updatedApp, scope),
	}, nil
}

func (s *AppServer) DeleteApp(ctx context.Context, req *pb.DeleteAppRequest) (*pb.DeleteAppResponse, error) {
	log.Println("Delete App")

	scope := PbScopeToOpScope(req.GetScope())
	auth, _, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	appID := req.GetId()
	err = s.AppStore.DeleteApp(appID)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to delete app")
	}

	return &pb.DeleteAppResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "App deleted successfully",
		},
	}, nil
}

func (s *AppServer) ApproveApp(ctx context.Context, req *pb.ApproveAppRequest) (*pb.ApproveAppResponse, error) {
	log.Println("Approve App")
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, "USERS")
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	app := PbAppToApp(req.GetApp(), &OpScope{
		Owner: OwnerType(pb.OwnerType_name[int32(req.GetApp().GetOwnerType())]),
		ID:    req.GetApp().GetOwnerId(),
	})
	app.ApprovedBy = contextData.UserID.String()
	app.ApprovedAt = time.Now()
	app.Status = k.GENERAL_STATUSES()["ACTIVE"]

	updatedApp, err := s.AppStore.UpdateApp(app)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to approve app")
	}

	updatedApp.Secret = "we do not store the app secret, Please check your email and keep it safe and secure."

	return &pb.ApproveAppResponse{
		App: AppToPbApp(updatedApp, &OpScope{
			updatedApp.OwnerType,
			updatedApp.OwnerId,
		}),
	}, nil
}

func (s *AppServer) RejectApp(ctx context.Context, req *pb.ApproveAppRequest) (*pb.ApproveAppResponse, error) {
	log.Println("Reject App")

	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, "USERS")
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	app := PbAppToApp(req.GetApp(), &OpScope{
		Owner: OwnerType(USERS),
		ID:    uuid.Nil.String(),
	})
	app.RejectedBy = contextData.UserID.String()
	app.RejectedAt = time.Now()
	app.Status = k.GENERAL_STATUSES()["REJECTED"]

	updatedApp, err := s.AppStore.UpdateApp(app)
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to reject app")
	}
	updatedApp.Secret = "we do not store the app secret, Please check your email and keep it safe and secure."

	return &pb.ApproveAppResponse{
		App: AppToPbApp(updatedApp, &OpScope{
			updatedApp.OwnerType,
			updatedApp.OwnerId,
		}),
	}, nil
}

func (s *AppServer) CreateTemplateApp(ctx context.Context, req *pb.CreateTemplateAppRequest) (*pb.CreateTemplateAppResponse, error) {
	log.Println("Create Template App")

	scope := PbScopeToOpScope(req.GetScope())
	auth, contextData, err := s.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	app := PbAppToApp(req.GetApp(), &OpScope{
		Owner: OwnerType(USERS),
		ID:    uuid.Nil.String(),
	})
	app.ID = uuid.New()
	app.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["apps_templates"], app.ID.String())
	app.Status = "TEMPLATE"
	app.CreatedBy = contextData.UserID.String()
	app.UpdatedBy = contextData.UserID.String()
	app.Token = "template"
	app.Secret = "template"
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	app.IsTemplate = true
	app.OwnerType = USERS
	app.OwnerId = "USERS"

	newApp, err := s.AppStore.SaveApp(app)
	if err != nil {
		return nil, status.Errorf(500, "Internal Server Error: Failed to create template app %v", err)
	}
	newApp.Secret = "We do not store the app secret, Please check your email and keep it safe and secure."

	return &pb.CreateTemplateAppResponse{
		App: AppToPbApp(newApp, &OpScope{
			newApp.OwnerType,
			newApp.OwnerId,
		}),
	}, nil
}

func (s *AppServer) ReadAppTemplates(ctx context.Context, req *pb.ReadAppTemplatesRequest) (*pb.ReadAppTemplatesResponse, error) {
	log.Println("Read App Templates")

	apps, err := s.AppStore.FindallTemplates()
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to get app templates")
	}

	return &pb.ReadAppTemplatesResponse{
		AppTemplates: AppsToPbApps(apps, &OpScope{
			Owner: OwnerType(USERS),
			ID:    uuid.Nil.String(),
		}),
	}, nil
}

func (s *AppServer) ReadConsoleApps(ctx context.Context, req *pb.ReadConsoleAppsRequest) (*pb.ReadConsoleAppsResponse, error) {
	log.Println("Read Console Apps")

	apps, err := s.AppStore.FindAll()
	if err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to get console apps")
	}

	return &pb.ReadConsoleAppsResponse{
		ConsoleApps: AppsToPbApps(apps, &OpScope{
			Owner: OwnerType(USERS),
			ID:    uuid.Nil.String(),
		}),
	}, nil
}
