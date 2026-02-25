package service

import (
	"context"
	"errors"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthStore struct {
	RbacStore         *RbacStore
	OrganisationStore *OrganisationStore
	JWTManager        *JWTManager
	AppStore          *AppStore
	BranchStore       *BranchStore
	UserStore         *UserStore
	RedisClient       *redis.Client
}

type IdNameMapping struct {
	ID   string
	Name string
}

type RequestMetadata struct {
	Authorization  string
	OrganisationID uuid.UUID
	UserID         uuid.UUID
	BranchID       uuid.UUID
	RequestAuth    string
	Roles          []uuid.UUID
	Scopes         *Scopes
	IdNameMappings []*IdNameMapping
	User           *User
}

type Scopes struct {
	User         uuid.UUID
	Teams        []string
	Departments  []string
	Branches     []string
	Branch       uuid.UUID // Current Branch
	Organisation uuid.UUID
}

type ApiAppMetadata struct {
	OrganisationID uuid.UUID
	AppID          uuid.UUID
	Token          string
	RequestAuth    string
}

func NewAuthStore(
	RbacStore *RbacStore,
	OrganisationStore *OrganisationStore,
	JWTManager *JWTManager,
	AppStore *AppStore,
	BranchStore *BranchStore,
	UserStore *UserStore,
	RedisClient *redis.Client,
) *AuthStore {
	return &AuthStore{
		RbacStore:         RbacStore,
		OrganisationStore: OrganisationStore,
		JWTManager:        JWTManager,
		AppStore:          AppStore,
		BranchStore:       BranchStore,
		UserStore:         UserStore,
		RedisClient:       RedisClient,
	}
}

func (s *AuthStore) AuthInternalRequests(token string, permissionCodeName string) (*RequestMetadata, error) {
	claims, err := s.JWTManager.ValidateToken(token)
	if err != nil {
		return nil, errors.New("could not Validate User Access")
	}
	if claims == nil {
		return nil, errors.New("invalid token claims")
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, status.Errorf(401, "Token Expired")
	}

	requestAuth := ""
	orgId := uuid.Nil
	userId := uuid.Nil
	branchId := uuid.Nil
	roles := []uuid.UUID{uuid.Nil}
	scopes := &Scopes{}
	idNameMappings := []*IdNameMapping{}

	if claims.UserId != uuid.Nil {
		userId = claims.UserId
	}

	if claims.BranchID != uuid.Nil {
		branchId = claims.BranchID
	}

	if claims.OrganisationID != uuid.Nil {
		orgId = claims.OrganisationID
	}

	if len(claims.Roles) > 0 {
		for _, role := range claims.Roles {
			roles = append(roles, uuid.MustParse(role))
		}
	}

	if permissionCodeName == "" {
		requestAuth = k.NewConsts().TRUE
	} else {
		roles := []uuid.UUID{}
		for _, role := range claims.Roles {
			roles = append(roles, uuid.MustParse(role))
		}
		scopes, requestAuth, idNameMappings = s.UserRequestAuthorization(roles, permissionCodeName, claims.UserId)
	}

	return &RequestMetadata{
		Authorization:  token,
		OrganisationID: orgId,
		UserID:         userId,
		BranchID:       branchId,
		RequestAuth:    requestAuth,
		Roles:          roles,
		Scopes:         scopes,
		IdNameMappings: idNameMappings,
	}, nil
}

func (s *AuthStore) AuthRequestAuth(ctx context.Context, permissionCodeName string) (*RequestMetadata, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		// Fallback for streaming calls
		if streamMD, ok := metadata.FromIncomingContext(context.Background()); ok {
			md = streamMD
		} else {
			return nil, errors.New("metadata not found in any context")
		}
	}

	requestMetadata := &RequestMetadata{}
	scopes := &Scopes{}
	idNameMappings := []*IdNameMapping{}
	auth := "false"
	if authorizationHeader, ok := md["authorization"]; ok {
		claims, err := s.JWTManager.ValidateToken(authorizationHeader[0])
		roles := []uuid.UUID{}
		if err != nil {
			return nil, status.Errorf(401, fmt.Sprintf("could not Validate User Access: %s", err.Error()))
		}
		if claims == nil {
			return nil, status.Errorf(401, "could not Validate User Access: invalid token claims")
		}
		if claims.ExpiresAt < time.Now().Unix() {
			return nil, status.Errorf(401, "token is expired")
		}
		for _, role := range claims.Roles {
			r, err := uuid.Parse(role)
			if err != nil {
				continue
			}
			roles = append(roles, r)
		}
		if claims.UserId != uuid.Nil {
			requestMetadata.UserID = claims.UserId
		}

		if len(authorizationHeader[0]) > 0 {
			requestMetadata.Authorization = authorizationHeader[0]
		}
		if len(claims.Roles) > 0 {
			requestMetadata.Roles = roles
		}
		if claims.OrganisationID != uuid.Nil {
			requestMetadata.OrganisationID = claims.OrganisationID
		}
		if claims.BranchID != uuid.Nil {
			requestMetadata.BranchID = claims.BranchID
		}
		if permissionCodeName == "" {
			auth = k.NewConsts().TRUE
		} else {
			scopes, auth, idNameMappings = s.UserRequestAuthorization(roles, permissionCodeName, claims.UserId)
		}
	} else if apikey, ok := md["apikey"]; ok {
		app, err := s.ValidateKeyAndSecret(md["apikey"][0], md["secret"][0])
		if err != nil {
			return nil, status.Errorf(401, "could not Validate App Access: %s", err.Error())
		}
		if app.Status != k.GENERAL_STATUSES()["ACTIVE"] {
			return nil, status.Errorf(401, "App is not active")
		}
		if app.ID != uuid.Nil {
			requestMetadata.UserID = app.ID
		}
		if len(apikey[0]) > 0 {
			requestMetadata.Authorization = app.Token
		}
		if app.OwnerType == OwnerType(BRANCHES) {
			requestMetadata.BranchID = uuid.MustParse(app.OwnerId)
		} else {
			requestMetadata.OrganisationID = uuid.MustParse(app.OwnerId)
		}
		if len(app.PermissionCodeNames) != 0 {
			auth = s.AppRequestAuthorization(app.PermissionCodeNames, permissionCodeName)
		}
	}
	requestMetadata.Scopes = scopes
	requestMetadata.IdNameMappings = idNameMappings
	requestMetadata.RequestAuth = auth
	return requestMetadata, nil
}

func (s *AuthStore) UserRequestAuthorization(roles []uuid.UUID, codeName string, user uuid.UUID) (scopes *Scopes, auth string, idNameMappings []*IdNameMapping) {
	userRoles, err := s.RbacStore.FindRolesByIDs(roles)
	if err != nil {
		return nil, k.NewConsts().FALSE, nil
	}
	teamPerms := []string{}
	departmentPerms := []string{}
	branchID := uuid.Nil
	orgID := uuid.Nil
	userID := uuid.Nil

	a := false
	if utils.ValueExistsGeneric(k.ROUTE_PERMISSIONS(), codeName) {
		for _, role := range userRoles {
			for _, perm := range role.PermissionCodeNames {
				if perm != codeName {
					continue
				}
				a = true
				switch role.OwnerType {
				case OwnerType(TEAMS):
					teamPerms = append(teamPerms, role.OwnerID.String())
					teamMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Team{}, role.OwnerID.String(), TeamNameExtractor)
					if err != nil {
						log.Printf("Error finding team mapping: %v", err)
					}
					idNameMappings = append(idNameMappings, &teamMapping)
				case OwnerType(DEPARTMENTS):
					departmentPerms = append(departmentPerms, role.OwnerID.String())
					departmentMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Department{}, role.OwnerID.String(), DepartmentNameExtractor)
					if err != nil {
						log.Printf("Error finding department mapping: %v", err)
					}
					idNameMappings = append(idNameMappings, &departmentMapping)
				case OwnerType(BRANCHES):
					branchID = role.OwnerID
					branchMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Branch{}, branchID.String(), BranchNameExtractor)
					if err != nil {
						log.Printf("Error finding branch mapping: %v", err)
					}
					idNameMappings = append(idNameMappings, &branchMapping)
					userID = user
					userMapping, err := s.RbacStore.FindDynamicIdNameMapping(&User{}, user.String(), UserNameExtractor)
					if err != nil {
						log.Printf("Error finding user mapping: %v", err)
					}
					idNameMappings = append(idNameMappings, &userMapping)
				case OwnerType(ORGANISATIONS):
					orgID = role.OwnerID
					orgMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Organisation{}, orgID.String(), OrganisationNameExtractor)
					if err != nil {
						log.Printf("Error finding organisation mapping: %v", err)
					}
					idNameMappings = append(idNameMappings, &orgMapping)
				}
			}
		}
	}
	log.Printf("assigning scope owners from roles %v", codeName)
	if utils.ValueExistsGeneric(k.OPEN_ROUTES(), codeName) {
		for _, role := range userRoles {
			switch role.OwnerType {
			case OwnerType(TEAMS):
				teamPerms = append(teamPerms, role.OwnerID.String())
				teamMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Team{}, role.OwnerID.String(), TeamNameExtractor)
				if err != nil {
					log.Printf("Error finding team mapping: %v", err)
				}
				idNameMappings = append(idNameMappings, &teamMapping)
			case OwnerType(DEPARTMENTS):
				departmentPerms = append(departmentPerms, role.OwnerID.String())
				departmentMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Department{}, role.OwnerID.String(), DepartmentNameExtractor)
				if err != nil {
					log.Printf("Error finding department mapping: %v", err)
				}
				idNameMappings = append(idNameMappings, &departmentMapping)
			case OwnerType(BRANCHES):
				branchID = role.OwnerID
				branchMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Branch{}, branchID.String(), BranchNameExtractor)
				if err != nil {
					log.Printf("Error finding branch mapping: %v", err)
				}
				idNameMappings = append(idNameMappings, &branchMapping)
				userID = user
				userMapping, err := s.RbacStore.FindDynamicIdNameMapping(&User{}, user.String(), UserNameExtractor)
				if err != nil {
					log.Printf("Error finding user mapping: %v", err)
				}
				idNameMappings = append(idNameMappings, &userMapping)
			case OwnerType(ORGANISATIONS):
				orgID = role.OwnerID
				orgMapping, err := s.RbacStore.FindDynamicIdNameMapping(&Organisation{}, orgID.String(), OrganisationNameExtractor)
				if err != nil {
					log.Printf("Error finding organisation mapping: %v", err)
				}
				idNameMappings = append(idNameMappings, &orgMapping)
			}
		}
		a = true
	}
	scope := &Scopes{
		Teams:        teamPerms,
		Departments:  departmentPerms,
		Branch:       branchID,
		Organisation: orgID,
		User:         userID,
	}
	return scope, k.NewConsts().BoolToString(a), idNameMappings
}

func (s *AuthStore) AppRequestAuthorization(permissions []string, codeName string) string {
	for _, perm := range permissions {
		if perm == codeName {
			return k.NewConsts().TRUE
		}
	}
	return k.NewConsts().FALSE
}

func (s *AuthStore) GetServiceAuthMetadata(ctx context.Context) (*RequestMetadata, error) {
	data := RequestMetadata{
		UserID:         uuid.Nil,
		OrganisationID: uuid.Nil,
		BranchID:       uuid.Nil,
		RequestAuth:    "",
		Scopes:         &Scopes{},
	}
	if userID, ok := ctx.Value(UserID).(uuid.UUID); ok {
		data.UserID = userID
	}
	if organisationID, ok := ctx.Value(OrganisationID).(uuid.UUID); ok {
		data.OrganisationID = organisationID
	}
	if branchID, ok := ctx.Value(BranchID).(uuid.UUID); ok {
		data.BranchID = branchID
	}
	if requestAuth, ok := ctx.Value(RequestAuth).(string); ok {
		data.RequestAuth = requestAuth
	}
	if authorization, ok := ctx.Value(Authorization).(string); ok {
		data.Authorization = authorization
	}
	if scope, ok := ctx.Value(Scope).(*Scopes); ok {
		data.Scopes = scope
	}
	return &data, nil
}

func (s *AuthStore) GetUserRequestMetadata(ctx context.Context, dataChan chan *RequestMetadata, err chan error) {
	data := RequestMetadata{
		UserID:         uuid.Nil,
		OrganisationID: uuid.Nil,
		BranchID:       uuid.Nil,
		RequestAuth:    "",
		Authorization:  "",
		Scopes:         &Scopes{},
	}
	// retrieve already set values from the context.Value with keys coresponding to data keys and assert the values to be of type string. if it is not, assign empty string
	if userID, ok := ctx.Value(UserID).(uuid.UUID); ok {
		data.UserID = userID
	}
	if organisationID, ok := ctx.Value(OrganisationID).(uuid.UUID); ok {
		data.OrganisationID = organisationID
	}
	if branchID, ok := ctx.Value(BranchID).(uuid.UUID); ok {
		data.BranchID = branchID
	}
	if requestAuth, ok := ctx.Value(RequestAuth).(string); ok {
		data.RequestAuth = requestAuth
	}
	if authorization, ok := ctx.Value(Authorization).(string); ok {
		data.Authorization = authorization
	}
	if scope, ok := ctx.Value(Scope).(*Scopes); ok {
		data.Scopes = scope
	}

	dataChan <- &data
	err <- nil
}

func (s *AuthStore) ReadContextOrganisation(organisationID uuid.UUID) (*Organisation, error) {
	org, err := s.OrganisationStore.FindByID(&organisationID)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (s *AuthStore) ValidateKeyAndSecret(apikey string, secret string) (*App, error) {
	app, err := s.AppStore.FindByToken(apikey)
	if err != nil {
		return nil, err
	}
	if app.Status != k.GENERAL_STATUSES()["ACTIVE"] {
		return nil, errors.New("App is not active")
	}

	if !utils.COMPARE_PASSWORD(app.Secret, secret) {
		return nil, errors.New("invalid api key")
	}
	return app, nil
}

func (s *AuthStore) InternalScopeCheck(contextData *RequestMetadata, req string) (*RequestMetadata, error) {
	permissionScopes, auth, idNameMappings := s.UserRequestAuthorization(contextData.Roles, req, contextData.UserID)
	permissionScopes.User = contextData.UserID
	metadata := &RequestMetadata{
		RequestAuth:    auth,
		Authorization:  contextData.Authorization,
		OrganisationID: contextData.OrganisationID,
		UserID:         contextData.UserID,
		BranchID:       contextData.BranchID,
		Roles:          contextData.Roles,
		Scopes:         permissionScopes,
		IdNameMappings: idNameMappings,
	}

	return metadata, nil
}

func (s *AuthStore) ScopeCheck(ctx context.Context, scopeID string) (bool, *RequestMetadata, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go s.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return false, nil, status.Error(403, "Forbidden, You do not have access to this resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scopeID) {
			utils.LogAsJSON("scope", contextChanData.Scopes)
			return false, nil, status.Error(403, "Forbidden, You do not have access to this resource")
		}
		return true, contextChanData, nil
	case err := <-errChan:
		return false, nil, status.Errorf(500, "error while fetching user %v", err)
	}
}

func GetIPAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}

	addr := p.Addr.String() // e.g. "192.168.1.10:54321"

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr // fallback, just return raw addr
	}

	return ip
}

// MARK: Session CRUD

type DBAuthStore struct {
	DB *gorm.DB
}

func NewDBAuthStore(db *gorm.DB) *DBAuthStore {
	return &DBAuthStore{
		DB: db,
	}
}

func (s *DBAuthStore) CreateSession(session *UserSession) (*UserSession, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(session).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *DBAuthStore) InvalidateSession(id string) error {
	session := &UserSession{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("id = ?", id).First(session)
		session.IsValid = false
		session.EndTime = time.Now().Format(time.RFC3339)
		return tx.Save(session).Error
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *DBAuthStore) GetSession(id string) (*UserSession, error) {
	session := &UserSession{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(session).Error
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *DBAuthStore) GetActiveSessions(userID uuid.UUID) ([]*UserSession, error) {
	sessions := []*UserSession{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("user_id = ? AND is_valid = ?", userID, true).Find(&sessions).Error
	})
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *DBAuthStore) GetAllSessions(scope *OpScope) ([]*UserSession, error) {
	sessions := []*UserSession{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("is_valid = ? AND owner_id = ? AND owner_type = ?", true, scope.ID, scope.Owner).Find(&sessions).Error
	})
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Create User Devices
func (s *DBAuthStore) CreateDevice(device *UserDevice) (*UserDevice, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Save(device).Error
	})
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s *DBAuthStore) GetDevice(id string) (*UserDevice, error) {
	device := &UserDevice{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(device).Error
	})
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s *DBAuthStore) GetUserDevices(userID uuid.UUID) ([]*UserDevice, error) {
	devices := []*UserDevice{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("user_id = ?", userID).Find(&devices).Error
	})
	if err != nil {
		return nil, err
	}
	return devices, nil
}
