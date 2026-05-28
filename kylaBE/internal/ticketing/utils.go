package ticketing

import "kyla-be/pkg/pb"

// RoomToPb converts a TicketRoom model to its proto representation.
func RoomToPb(r *TicketRoom) *pb.TicketRoom {
	return &pb.TicketRoom{
		Id:           r.ID,
		TicketId:     r.TicketID,
		OrgId:        r.OrgID,
		Name:         r.Name,
		Type:         roomTypeStringToPb(r.Type),
		MessageCount: int32(r.MessageCount),
		CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// MessageToPb converts a TicketRoomMessage model to its proto representation.
func MessageToPb(m *TicketRoomMessage) *pb.TicketRoomMessage {
	return &pb.TicketRoomMessage{
		Id:        m.ID,
		RoomId:    m.RoomID,
		OrgId:     m.OrgID,
		AuthorId:  m.AuthorID,
		Content:   m.Content,
		IsPrivate: m.IsPrivate,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// MacroToPb converts a Macro model to its proto representation.
func MacroToPb(m *Macro) *pb.Macro {
	return &pb.Macro{
		Id:          m.ID,
		OrgId:       m.OrgID,
		WorkspaceId: m.WorkspaceID,
		Name:        m.Name,
		Content:     m.Content,
		Actions:     m.Actions,
		Visibility:  macroVisibilityStringToPb(m.Visibility),
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func roomTypeStringToPb(t string) pb.RoomType {
	switch t {
	case "customer_reply":
		return pb.RoomType_ROOM_TYPE_CUSTOMER_REPLY
	default:
		return pb.RoomType_ROOM_TYPE_INTERNAL
	}
}

func roomTypeFromPb(t pb.RoomType) string {
	switch t {
	case pb.RoomType_ROOM_TYPE_CUSTOMER_REPLY:
		return "customer_reply"
	default:
		return "internal"
	}
}

func macroVisibilityStringToPb(v string) pb.MacroVisibility {
	switch v {
	case "team":
		return pb.MacroVisibility_MACRO_VISIBILITY_TEAM
	case "public":
		return pb.MacroVisibility_MACRO_VISIBILITY_PUBLIC
	default:
		return pb.MacroVisibility_MACRO_VISIBILITY_PRIVATE
	}
}

func macroVisibilityFromPb(v pb.MacroVisibility) string {
	switch v {
	case pb.MacroVisibility_MACRO_VISIBILITY_TEAM:
		return "team"
	case pb.MacroVisibility_MACRO_VISIBILITY_PUBLIC:
		return "public"
	default:
		return "private"
	}
}
