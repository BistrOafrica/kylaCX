package authctx

import "kyla-be/pkg/pb"

// PbScopeToOpScope converts a protobuf Scope to an OpScope.
func PbScopeToOpScope(scope *pb.Scope) *OpScope {
	return &OpScope{
		Owner: OwnerType(pb.OwnerType_name[int32(scope.OwnerType)]),
		ID:    scope.OwnerId,
	}
}
