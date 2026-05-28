package auth

import (
	"context"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/service"

	"github.com/google/uuid"
)

// AuthStore wraps service.AuthStore and adapts methods to return *authctx.RequestMetadata
// instead of *service.RequestMetadata, preventing import cycles in domain packages.
type AuthStore struct {
	inner     *service.AuthStore
	rbacStore *rbac.RbacStore
}

// NewAuthStore wraps a *service.AuthStore for use in middleware and domain servers.
// rbacStore is the domain-package RbacStore needed by some domain server AuthGateway interfaces.
func NewAuthStore(s *service.AuthStore, rbacStore *rbac.RbacStore) *AuthStore {
	return &AuthStore{inner: s, rbacStore: rbacStore}
}

// Inner returns the underlying *service.AuthStore for cases that need full access.
func (a *AuthStore) Inner() *service.AuthStore {
	return a.inner
}

// GetRbacStore returns the domain rbac.RbacStore (satisfies team/department/etc AuthGateway).
func (a *AuthStore) GetRbacStore() *rbac.RbacStore {
	return a.rbacStore
}

// GetUserStore returns the service-layer UserStore (satisfies team AuthGateway).
func (a *AuthStore) GetUserStore() *service.UserStore {
	return a.inner.UserStore
}

// AuthRequestAuth delegates to the underlying service store and converts
// the result to *authctx.RequestMetadata.
func (a *AuthStore) AuthRequestAuth(ctx context.Context, permissionCodeName string) (*authctx.RequestMetadata, error) {
	svcMeta, err := a.inner.AuthRequestAuth(ctx, permissionCodeName)
	if err != nil {
		return nil, err
	}
	return convertToAuthCtxMetadata(svcMeta), nil
}

// GetServiceAuthMetadata reads auth context values injected by the middleware interceptor
// using the authctx package keys. This overrides delegation to a.inner.GetServiceAuthMetadata
// which uses service-package key constants that are a different Go type — ctx.Value(service.UserID)
// cannot find values set by the middleware with ctx.Value(authctx.UserID).
func (a *AuthStore) GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error) {
	data := &authctx.RequestMetadata{
		Scopes: &authctx.Scopes{},
	}
	if v, ok := ctx.Value(authctx.UserID).(uuid.UUID); ok {
		data.UserID = v
	}
	if v, ok := ctx.Value(authctx.OrganisationID).(uuid.UUID); ok {
		data.OrganisationID = v
	}
	if v, ok := ctx.Value(authctx.BranchID).(uuid.UUID); ok {
		data.BranchID = v
	}
	if v, ok := ctx.Value(authctx.WorkspaceID).(uuid.UUID); ok {
		data.WorkspaceID = v
	}
	if v, ok := ctx.Value(authctx.RequestAuth).(string); ok {
		data.RequestAuth = v
	}
	if v, ok := ctx.Value(authctx.Authorization).(string); ok {
		data.Authorization = v
	}
	if v, ok := ctx.Value(authctx.Scope).(*authctx.Scopes); ok && v != nil {
		data.Scopes = v
	}
	return data, nil
}

// GetUserRequestMetadata delegates to the underlying service store, adapting the channel types.
func (a *AuthStore) GetUserRequestMetadata(ctx context.Context, metaChan chan *authctx.RequestMetadata, errChan chan error) {
	svcMetaChan := make(chan *service.RequestMetadata, 1)
	go a.inner.GetUserRequestMetadata(ctx, svcMetaChan, errChan)
	go func() {
		if m, ok := <-svcMetaChan; ok {
			metaChan <- convertToAuthCtxMetadata(m)
		}
	}()
}

// ScopeCheck delegates to the underlying service store and converts the result.
func (a *AuthStore) ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error) {
	ok, svcMeta, err := a.inner.ScopeCheck(ctx, scopeID)
	if err != nil {
		return false, nil, err
	}
	return ok, convertToAuthCtxMetadata(svcMeta), nil
}

// convertToAuthCtxMetadata converts a *service.RequestMetadata to *authctx.RequestMetadata.
func convertToAuthCtxMetadata(s *service.RequestMetadata) *authctx.RequestMetadata {
	if s == nil {
		return nil
	}

	var scopes *authctx.Scopes
	if s.Scopes != nil {
		scopes = &authctx.Scopes{
			User:         s.Scopes.User,
			Teams:        s.Scopes.Teams,
			Departments:  s.Scopes.Departments,
			Branches:     s.Scopes.Branches,
			Branch:       s.Scopes.Branch,
			Organisation: s.Scopes.Organisation,
		}
	}

	idNameMappings := make([]*authctx.IdNameMapping, 0, len(s.IdNameMappings))
	for _, m := range s.IdNameMappings {
		if m != nil {
			idNameMappings = append(idNameMappings, &authctx.IdNameMapping{
				ID:   m.ID,
				Name: m.Name,
			})
		}
	}

	return &authctx.RequestMetadata{
		Authorization:  s.Authorization,
		OrganisationID: s.OrganisationID,
		UserID:         s.UserID,
		BranchID:       s.BranchID,
		RequestAuth:    s.RequestAuth,
		Roles:          s.Roles,
		Scopes:         scopes,
		IdNameMappings: idNameMappings,
	}
}
