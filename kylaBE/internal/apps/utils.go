package apps

import (
	authctx "kyla-be/internal/authctx"
	"kyla-be/pkg/pb"

	"github.com/google/uuid"
)

func PbAppToApp(pbApp *pb.App, scope *authctx.OpScope) *App {
	id, err := uuid.Parse(pbApp.GetId())
	if err != nil {
		id = uuid.New()
	}

	return &App{
		ID:                  id,
		Token:               pbApp.GetToken(),
		Name:                pbApp.GetName(),
		Description:         pbApp.GetDescription(),
		Secret:              pbApp.GetSecret(),
		Status:              pbApp.GetStatus(),
		CreatedBy:           pbApp.GetCreatedBy(),
		UpdatedBy:           pbApp.GetUpdatedBy(),
		ApprovedBy:          pbApp.GetApprovedBy(),
		RejectedBy:          pbApp.GetRejectedBy(),
		PermissionCodeNames: pbApp.GetPermissionsCodenames(),
		IsTemplate:          false,
		OwnerType:           scope.Owner,
		OwnerId:             scope.ID,
	}
}

func AppToPbApp(app *App, scope *authctx.OpScope) *pb.App {
	return &pb.App{
		Id:                   app.ID.String(),
		Token:                app.Token,
		Name:                 app.Name,
		Description:          app.Description,
		Secret:               app.Secret,
		Status:               app.Status,
		CreatedBy:            app.CreatedBy,
		CreatedAt:            app.CreatedAt.String(),
		UpdatedBy:            app.UpdatedBy,
		UpdatedAt:            app.UpdatedAt.String(),
		ApprovedBy:           app.ApprovedBy,
		RejectedBy:           app.RejectedBy,
		RejectedAt:           app.RejectedAt.String(),
		PermissionsCodenames: app.PermissionCodeNames,
		IsTemplate:           app.IsTemplate,
		OwnerType:            pb.OwnerType(pb.OwnerType_value[string(app.OwnerType)]),
		OwnerId:              app.OwnerId,
	}
}

func PbAppsToApps(pbApps []*pb.App, scope *authctx.OpScope) []*App {
	apps := make([]*App, 0)
	for _, pbApp := range pbApps {
		apps = append(apps, PbAppToApp(pbApp, scope))
	}
	return apps
}

func AppsToPbApps(apps []*App, scope *authctx.OpScope) []*pb.App {
	pbApps := make([]*pb.App, 0)
	for _, app := range apps {
		pbApps = append(pbApps, AppToPbApp(app, scope))
	}
	return pbApps
}
