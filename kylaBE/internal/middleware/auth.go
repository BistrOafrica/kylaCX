package middleware

import (
	"context"
	"fmt"
	"log"

	"kyla-be/internal/auth"
	"kyla-be/internal/authctx"
	casbinsvc "kyla-be/internal/casbin"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor validates JWT tokens and enforces Casbin RBAC on every gRPC call.
type AuthInterceptor struct {
	jwtManager *auth.JWTManager
	authStore  *auth.AuthStore
	enforcer   *casbinsvc.Enforcer
}

// NewAuthInterceptor creates a new AuthInterceptor with Casbin enforcement.
func NewAuthInterceptor(jwtManager *auth.JWTManager, authStore *auth.AuthStore, enforcer *casbinsvc.Enforcer) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
		authStore:  authStore,
		enforcer:   enforcer,
	}
}

// Unary returns a server interceptor for authentication and authorization.
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		// ReadAuthContext is an internal no-auth path — pass through directly.
		if info.FullMethod == "/da.proto.AuthService/ReadAuthContext" {
			return handler(ctx, req)
		}

		// Validate JWT and extract user claims.
		// Empty permissionCodeName skips the legacy UserRequestAuthorization GORM query.
		requestMetadata, err := interceptor.authStore.AuthRequestAuth(ctx, "")
		if err != nil {
			log.Println("Error while validating token:", err)
			return nil, err
		}
		if requestMetadata == nil {
			return nil, status.Errorf(500, "internal server error: requestMetadata is nil")
		}

		// Extract x-workspace-id from gRPC incoming metadata header.
		var workspaceID uuid.UUID
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-workspace-id"); len(vals) > 0 {
				if wsID, parseErr := uuid.Parse(vals[0]); parseErr == nil {
					workspaceID = wsID
				}
			}
		}

		// Determine request authorisation via Casbin RBAC.
		requestAuth := interceptor.casbinAuth(info.FullMethod, requestMetadata, workspaceID)

		// Inject all context values for downstream domain handlers.
		ctx = context.WithValue(ctx, authctx.UserID, requestMetadata.UserID)
		ctx = context.WithValue(ctx, authctx.OrganisationID, requestMetadata.OrganisationID)
		ctx = context.WithValue(ctx, authctx.Authorization, requestMetadata.Authorization)
		ctx = context.WithValue(ctx, authctx.RequestAuth, requestAuth)
		ctx = context.WithValue(ctx, authctx.BranchID, requestMetadata.BranchID)
		ctx = context.WithValue(ctx, authctx.Roles, requestMetadata.Roles)
		ctx = context.WithValue(ctx, authctx.Scope, requestMetadata.Scopes)
		ctx = context.WithValue(ctx, authctx.WorkspaceID, workspaceID)

		return handler(ctx, req)
	}
}

// casbinAuth returns k.NewConsts().TRUE or FALSE based on the Casbin policy check.
// Open routes pass through if JWT is valid. Unknown routes are denied by default.
func (interceptor *AuthInterceptor) casbinAuth(fullMethod string, meta *authctx.RequestMetadata, workspaceID uuid.UUID) string {
	consts := k.NewConsts()

	// ReadMe is always allowed for any authenticated user.
	if fullMethod == "/da.proto.UserService/ReadMe" {
		return consts.TRUE
	}

	policy, found := casbinsvc.MethodPolicy(fullMethod)
	if !found {
		// Route not in map — deny by default (fail closed).
		return consts.FALSE
	}
	if policy.Open {
		// Open route — JWT validity is sufficient, no RBAC check needed.
		return consts.TRUE
	}

	if interceptor.enforcer == nil {
		// Enforcer not wired — deny to prevent accidental open access.
		return consts.FALSE
	}

	// Workspace domain takes precedence over org domain when provided.
	var domain string
	if workspaceID != uuid.Nil {
		domain = fmt.Sprintf("ws:%s", workspaceID.String())
	} else {
		domain = fmt.Sprintf("org:%s", meta.OrganisationID.String())
	}

	subject := fmt.Sprintf("user:%s", meta.UserID.String())
	allowed, err := interceptor.enforcer.Enforce(subject, domain, policy.Resource, policy.Action)
	if err != nil {
		log.Printf("[auth] casbin enforce error for %s: %v", fullMethod, err)
		return consts.FALSE
	}
	if allowed {
		return consts.TRUE
	}
	return consts.FALSE
}

// Stream returns a server interceptor for streaming calls.
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream interceptor: ", info.FullMethod)

		if !k.IsStreamRouteAuthenticated(info.FullMethod) {
			return handler(srv, stream)
		}

		ctx := stream.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Println("Stream interceptor: No metadata found")
		}

		if len(md) == 0 {
			if headers, ok := metadata.FromIncomingContext(stream.Context()); ok {
				md = headers
			}
		}

		if len(md) > 0 {
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		requestMetadata, err := interceptor.authStore.AuthRequestAuth(ctx, "")
		if err != nil {
			log.Println("Error: ", err)
			return err
		}

		// Extract x-workspace-id from stream metadata.
		var workspaceID uuid.UUID
		if vals := md.Get("x-workspace-id"); len(vals) > 0 {
			if wsID, parseErr := uuid.Parse(vals[0]); parseErr == nil {
				workspaceID = wsID
			}
		}

		requestAuth := interceptor.casbinAuth(info.FullMethod, requestMetadata, workspaceID)

		wrappedStream := &WrappedServerStream{
			ServerStream: stream,
			ctx:          ctx,
		}
		wrappedStream.ctx = context.WithValue(ctx, authctx.UserID, requestMetadata.UserID)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.OrganisationID, requestMetadata.OrganisationID)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.Authorization, requestMetadata.Authorization)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.RequestAuth, requestAuth)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.BranchID, requestMetadata.BranchID)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.Roles, requestMetadata.Roles)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.Scope, requestMetadata.Scopes)
		wrappedStream.ctx = context.WithValue(wrappedStream.ctx, authctx.WorkspaceID, workspaceID)

		return handler(srv, wrappedStream)
	}
}

// WrappedServerStream is a ServerStream with an overridable context.
// It is shared between auth and response interceptors.
type WrappedServerStream struct {
	grpc.ServerStream
	ctx              context.Context
	method           string
	shouldLog        bool
	sqsActions       *utils.SQSActions
	incomingMetadata metadata.MD
}

// Context returns the wrapped context.
func (w *WrappedServerStream) Context() context.Context {
	return w.ctx
}

// SessionDevicesInterceptors handles device/session validation per request.
type SessionDevicesInterceptors struct {
	DBAuthStore *auth.DBAuthStore
}

// NewSessionDevicesInterceptors creates a new SessionDevicesInterceptors.
func NewSessionDevicesInterceptors(dbAuthStore *auth.DBAuthStore) *SessionDevicesInterceptors {
	return &SessionDevicesInterceptors{
		DBAuthStore: dbAuthStore,
	}
}

// Unary returns the device session unary interceptor.
func (s *SessionDevicesInterceptors) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)
		// TODO: Devices registration.
		return handler(ctx, req)
	}
}
