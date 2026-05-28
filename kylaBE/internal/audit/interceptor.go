package audit

import (
	"context"
	"log"
	"net"

	"kyla-be/internal/authctx"
	casbinroutes "kyla-be/internal/casbin"
	"kyla-be/pkg/k"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// Interceptor records every gRPC auth decision to the audit log.
// It must run AFTER the AuthInterceptor so that context values are already set.
type Interceptor struct {
	store *Store
}

// NewInterceptor returns an audit Interceptor.
func NewInterceptor(store *Store) *Interceptor {
	return &Interceptor{store: store}
}

// Unary returns a gRPC unary interceptor that asynchronously writes an AuditLog entry
// after each request. It never blocks or fails the RPC on logging errors.
func (i *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		go i.record(ctx, info.FullMethod, err == nil)
		return resp, err
	}
}

// Stream returns a gRPC stream interceptor that asynchronously records stream initiations.
func (i *Interceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := handler(srv, stream)
		go i.record(stream.Context(), info.FullMethod, err == nil)
		return err
	}
}

// record builds and persists an AuditLog entry. Runs in a goroutine.
func (i *Interceptor) record(ctx context.Context, fullMethod string, allowed bool) {
	entry := &AuditLog{
		ID:      uuid.New(),
		Method:  fullMethod,
		Allowed: allowed,
	}

	// Extract userID from context (set by AuthInterceptor).
	if uid, ok := ctx.Value(authctx.UserID).(uuid.UUID); ok {
		entry.UserID = uid.String()
	}

	// Extract orgID from context.
	if oid, ok := ctx.Value(authctx.OrganisationID).(uuid.UUID); ok {
		entry.OrgID = oid.String()
	}

	// Extract workspaceID from context.
	if wid, ok := ctx.Value(authctx.WorkspaceID).(uuid.UUID); ok && wid != uuid.Nil {
		entry.WorkspaceID = wid.String()
	}

	// Override allowed: if requestAuth in context is false, mark as denied.
	if ra, ok := ctx.Value(authctx.RequestAuth).(string); ok && ra == k.NewConsts().FALSE {
		entry.Allowed = false
	}

	// Look up Casbin resource/action for additional context.
	if rp, found := casbinroutes.MethodPolicy(fullMethod); found {
		entry.Resource = rp.Resource
		entry.Action = rp.Action
	}

	// Extract client IP if available.
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		addr := p.Addr.String()
		if host, _, err := net.SplitHostPort(addr); err == nil {
			entry.IPAddress = host
		} else {
			entry.IPAddress = addr
		}
	}

	if err := i.store.Create(entry); err != nil {
		log.Printf("[audit] failed to write log for %s: %v", fullMethod, err)
	}
}
