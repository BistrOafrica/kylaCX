package invitation

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// InvitationStore defines the interface for invitation data access
type InvitationStore struct {
	db *gorm.DB
}

// NewInvitationStore creates a new repository instance
func NewInvitationStore(db *gorm.DB) *InvitationStore {
	return &InvitationStore{db: db}
}

// Create inserts a new invitation into the database
func (r *InvitationStore) Create(ctx context.Context, invitation *Invitation) error {
	result := r.db.WithContext(ctx).Create(invitation)
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to create invitation")
	}
	return nil
}

// GetByID retrieves an invitation by its ID
func (r *InvitationStore) GetByID(ctx context.Context, id string) (*Invitation, error) {
	var invitation Invitation
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(result.Error, "failed to get invitation by ID")
	}
	return &invitation, nil
}

// GetByToken retrieves an invitation by its token
func (r *InvitationStore) GetByToken(ctx context.Context, token string) (*Invitation, error) {
	var invitation Invitation
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(result.Error, "failed to get invitation by token")
	}
	return &invitation, nil
}

// List retrieves a paginated list of invitations
func (r *InvitationStore) List(ctx context.Context, organisationID string, status InvitationStatus, page, pageSize int) ([]*Invitation, int, error) {
	offset := (page - 1) * pageSize

	// Build query
	query := r.db.WithContext(ctx).Model(&Invitation{}).Where("organisation_id = ?", organisationID)

	// Add status filter if specified
	if status != InvitationStatusUnspecified {
		query = query.Where("status = ?", status)
	}

	// Get total count
	var total int64
	countResult := query.Count(&total)
	if countResult.Error != nil {
		return nil, 0, errors.Wrap(countResult.Error, "failed to count invitations")
	}

	// Get paginated results
	var invitations []*Invitation
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&invitations)
	if result.Error != nil {
		return nil, 0, errors.Wrap(result.Error, "failed to list invitations")
	}

	return invitations, int(total), nil
}

// Update updates an existing invitation
func (r *InvitationStore) Update(ctx context.Context, invitation *Invitation) error {
	// Set updated_at to current time
	invitation.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Save(invitation)
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to update invitation")
	}

	if result.RowsAffected == 0 {
		return errors.New("invitation not found")
	}

	return nil
}

// Delete removes an invitation from the database
func (r *InvitationStore) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Invitation{}, "id = ?", id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to delete invitation")
	}

	if result.RowsAffected == 0 {
		return errors.New("invitation not found")
	}

	return nil
}
