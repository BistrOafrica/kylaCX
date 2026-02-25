package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is an interceptor for Authentication and Authorization.
type AuthInterceptor struct {
	jwtManager *JWTManager
	authStore  *AuthStore
}

// NewAuthInterceptor creates a new instance of AuthInterceptor.
func NewAuthInterceptor(jwtManager *JWTManager, authStore *AuthStore) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
		authStore:  authStore,
	}
}

type key int

const (
	UserID key = iota
	OrganisationID
	Authorization
	RequestAuth
	BranchID
	Roles
	Scope
)

// Unary returns a server interceptor for authentication and authorization.
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		var requestMetadata *RequestMetadata
		var err error

		if info.FullMethod == "/da.proto.AuthService/ReadAuthContext" {
			return handler(ctx, req)
		} else if ok := k.OPEN_ROUTES()[info.FullMethod]; ok != "" {
			requestMetadata, err = interceptor.authStore.AuthRequestAuth(ctx, "")
		} else if ok := k.ROUTE_PERMISSIONS()[info.FullMethod]; ok != "" {
			requestMetadata, err = interceptor.authStore.AuthRequestAuth(ctx, k.ROUTE_PERMISSIONS()[info.FullMethod])
		} else {
			requestMetadata = &RequestMetadata{}
		}
		if info.FullMethod == "/da.proto.UserService/ReadMe" {
			if err != nil {
				log.Println("Error while validating token:", err)
				return nil, err
			}
			if requestMetadata == nil {
				log.Println("Error: requestMetadata is nil")
				return nil, status.Errorf(500, "Internal server error: requestMetadata is nil")
			}
			ctx = context.WithValue(ctx, UserID, requestMetadata.UserID)
			ctx = context.WithValue(ctx, OrganisationID, requestMetadata.OrganisationID)
			ctx = context.WithValue(ctx, Authorization, requestMetadata.Authorization)
			ctx = context.WithValue(ctx, RequestAuth, k.NewConsts().TRUE)
			ctx = context.WithValue(ctx, BranchID, requestMetadata.BranchID)
			ctx = context.WithValue(ctx, Roles, requestMetadata.Roles)
			ctx = context.WithValue(ctx, Scope, requestMetadata.Scopes)
		} else if k.ROUTE_PERMISSIONS()[info.FullMethod] != "" || info.FullMethod == "/da.proto.AuthService/ReadAuthContext" || k.OPEN_ROUTES()[info.FullMethod] != "" {
			if err != nil {
				log.Println("Error: ", err)
				return nil, err
			}
			if requestMetadata == nil {
				log.Println("Error: requestMetadata is nil")
				return nil, status.Errorf(500, "Internal server error: requestMetadata is nil")
			}
			ctx = context.WithValue(ctx, UserID, requestMetadata.UserID)
			ctx = context.WithValue(ctx, OrganisationID, requestMetadata.OrganisationID)
			ctx = context.WithValue(ctx, Authorization, requestMetadata.Authorization)
			ctx = context.WithValue(ctx, RequestAuth, requestMetadata.RequestAuth)
			ctx = context.WithValue(ctx, BranchID, requestMetadata.BranchID)
			ctx = context.WithValue(ctx, Roles, requestMetadata.Roles)
			ctx = context.WithValue(ctx, Scope, requestMetadata.Scopes)
		} else {
			// For everything else, add the values but as empty with requestAuth as true
			ctx = context.WithValue(ctx, UserID, "")
			ctx = context.WithValue(ctx, OrganisationID, "")
			ctx = context.WithValue(ctx, Authorization, "")
			ctx = context.WithValue(ctx, RequestAuth, k.NewConsts().TRUE)
			ctx = context.WithValue(ctx, BranchID, "")
			ctx = context.WithValue(ctx, Roles, []string{})
			ctx = context.WithValue(ctx, Scope, []string{})
			ctx = context.WithValue(ctx, Scope, []string{})
		}
		return handler(ctx, req)
	}
}

// Stream returns a server interceptor for streaming calls for authentication and authorization.
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream interceptor: ", info.FullMethod)

		if k.IsStreamRouteAuthenticated(info.FullMethod) {
			ctx := stream.Context()
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				log.Println("Stream interceptor: No metadata found")
			}

			// Create new context with metadata if none exists
			if len(md) == 0 {
				// Try alternative metadata locations
				if headers, ok := metadata.FromIncomingContext(stream.Context()); ok {
					md = headers
				}
			}

			// Force metadata into context if needed
			if len(md) > 0 {
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			var requestMetadata *RequestMetadata
			var err error

			// Create a wrapped stream that will inject the context values
			wrappedStream := &WrappedServerStream{
				ServerStream: stream,
				ctx:          ctx,
			}

			if ok := k.ROUTE_PERMISSIONS()[info.FullMethod]; ok != "" {
				requestMetadata, err = interceptor.authStore.AuthRequestAuth(ctx, k.ROUTE_PERMISSIONS()[info.FullMethod])
			} else {
				requestMetadata = &RequestMetadata{}
			}

			if k.ROUTE_PERMISSIONS()[info.FullMethod] != "" || k.OPEN_ROUTES()[info.FullMethod] != "" {
				if err != nil {
					log.Println("Error: ", err)
					return err
				}
				wrappedStream.ctx = context.WithValue(ctx, UserID, requestMetadata.UserID)
				wrappedStream.ctx = context.WithValue(wrappedStream.ctx, OrganisationID, requestMetadata.OrganisationID)
				wrappedStream.ctx = context.WithValue(wrappedStream.ctx, Authorization, requestMetadata.Authorization)
				wrappedStream.ctx = context.WithValue(wrappedStream.ctx, RequestAuth, requestMetadata.RequestAuth)
				wrappedStream.ctx = context.WithValue(wrappedStream.ctx, BranchID, requestMetadata.BranchID)
				wrappedStream.ctx = context.WithValue(wrappedStream.ctx, Roles, requestMetadata.Roles)
				wrappedStream.ctx = context.WithValue(wrappedStream.ctx, Scope, requestMetadata.Scopes)
			}

			return handler(srv, wrappedStream)
		} else {
			return handler(srv, stream)
		}
	}
}

type WrappedServerStream struct {
	grpc.ServerStream
	ctx              context.Context
	method           string
	shouldLog        bool
	sqsActions       *utils.SQSActions
	incomingMetadata metadata.MD
}

// Context returns the overridden context
func (w *WrappedServerStream) Context() context.Context {
	return w.ctx
}

type SessionDevicesInterceptors struct {
	DBAuthStore *DBAuthStore
}

func NewSessionDevicesInterceptors(dbAuthStore *DBAuthStore) *SessionDevicesInterceptors {
	return &SessionDevicesInterceptors{
		DBAuthStore: dbAuthStore,
	}
}

func (s *SessionDevicesInterceptors) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)
		// TODO: Devices reegistration.
		return handler(ctx, req)
	}
}
