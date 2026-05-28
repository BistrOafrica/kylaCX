package workspace

import (
	"kyla-be/pkg/pb"
)

// WorkspaceToPb converts a Workspace model to its protobuf representation.
func WorkspaceToPb(w *Workspace) *pb.Workspace {
	if w == nil {
		return nil
	}
	return &pb.Workspace{
		Id:             w.ID,
		OrgId:          w.OrgID,
		Name:           w.Name,
		Slug:           w.Slug,
		Description:    w.Description,
		Icon:           w.Icon,
		Color:          w.Color,
		DomainTemplate: string(w.DomainTemplate),
		Status:         string(w.Status),
		CreatedAt:      w.CreatedAt.String(),
		UpdatedAt:      w.UpdatedAt.String(),
	}
}

// WorkspacesToPb converts a slice of Workspace models to their pb representations.
func WorkspacesToPb(workspaces []*Workspace) []*pb.Workspace {
	out := make([]*pb.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		out = append(out, WorkspaceToPb(w))
	}
	return out
}

// PbToWorkspace converts a pb.Workspace to a Workspace model.
func PbToWorkspace(p *pb.Workspace) *Workspace {
	if p == nil {
		return nil
	}
	return &Workspace{
		ID:             p.Id,
		OrgID:          p.OrgId,
		Name:           p.Name,
		Slug:           p.Slug,
		Description:    p.Description,
		Icon:           p.Icon,
		Color:          p.Color,
		DomainTemplate: DomainTemplate(p.DomainTemplate),
		Status:         WorkspaceStatus(p.Status),
	}
}

// MemberToPb converts a WorkspaceMember model to its protobuf representation.
func MemberToPb(m *WorkspaceMember) *pb.WorkspaceMember {
	if m == nil {
		return nil
	}
	return &pb.WorkspaceMember{
		WorkspaceId: m.WorkspaceID,
		UserId:      m.UserID,
		Role:        string(m.Role),
		JoinedAt:    m.JoinedAt.String(),
	}
}

// MembersToPb converts a slice of WorkspaceMember models.
func MembersToPb(members []*WorkspaceMember) []*pb.WorkspaceMember {
	out := make([]*pb.WorkspaceMember, 0, len(members))
	for _, m := range members {
		out = append(out, MemberToPb(m))
	}
	return out
}
