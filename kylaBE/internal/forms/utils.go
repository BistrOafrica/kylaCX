package forms

import "kyla-be/pkg/pb"

// FormToPb converts a Form model to its proto representation.
func FormToPb(f *Form) *pb.FormDefinition {
	return &pb.FormDefinition{
		Id:              f.ID,
		OrgId:           f.OrgID,
		WorkspaceId:     f.WorkspaceID,
		Name:            f.Name,
		Description:     f.Description,
		Fields:          f.Fields,
		Status:          formStatusStringToPb(f.Status),
		SubmitRedirect:  f.SubmitRedirect,
		SubmissionCount: int32(f.SubmissionCount),
		CreatedBy:       f.CreatedBy,
		CreatedAt:       f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// SubmissionToPb converts a FormSubmission to its proto representation.
func SubmissionToPb(s *FormSubmission) *pb.FormSubmission {
	sub := &pb.FormSubmission{
		Id:        s.ID,
		FormId:    s.FormID,
		OrgId:     s.OrgID,
		Data:      s.Data,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if s.ObjectID != nil {
		sub.ObjectId = s.ObjectID
	}
	return sub
}

func formStatusStringToPb(s string) pb.FormStatus {
	switch s {
	case "active":
		return pb.FormStatus_FORM_STATUS_ACTIVE
	case "closed":
		return pb.FormStatus_FORM_STATUS_CLOSED
	default:
		return pb.FormStatus_FORM_STATUS_DRAFT
	}
}

func formStatusFromPb(s pb.FormStatus) string {
	switch s {
	case pb.FormStatus_FORM_STATUS_ACTIVE:
		return "active"
	case pb.FormStatus_FORM_STATUS_CLOSED:
		return "closed"
	default:
		return "draft"
	}
}
