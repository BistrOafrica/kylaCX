package sharing

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SharingStore struct {
	db *gorm.DB
}

func NewSharingStore(db *gorm.DB) *SharingStore {
	return &SharingStore{db: db}
}

func (s *SharingStore) AddNode(node *Entity) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Save(node).Error
	})
	return err
}

func (s *SharingStore) AddEdge(edge *EntityLink) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Save(edge).Error
	})
	return err
}

func (s *SharingStore) GetResources(ownerID string) ([]EntityLink, error) {
	var edges []EntityLink
	if err := s.db.Preload(clause.Associations).Where("from_id = ?", ownerID).Find(&edges).Error; err != nil {
		return nil, err
	}
	return edges, nil
}

func (s *SharingStore) GetOwners(resourceID string) ([]EntityLink, error) {
	var edges []EntityLink
	if err := s.db.Preload(clause.Associations).Where("to_id = ?", resourceID).Find(&edges).Error; err != nil {
		return nil, err
	}
	return edges, nil
}

func (s *SharingStore) GetRequests(ownerID uuid.UUID) (*[]AccessRequest, error) {
	reqs := []AccessRequest{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		subQuery := tx.Table("entities").
			Select("id").
			Where("id = ? OR id IN (?)", ownerID,
				tx.Table("entity_links").
					Select("from_id").
					Where("to_id = ?", ownerID),
			)

		return tx.Where("resource_owner IN (?)", subQuery).Find(&reqs).Error
	})

	return &reqs, err
}

func (s *SharingStore) HasOwnership(ownerID, resourceID string) bool {
	var count int64
	s.db.Model(&EntityLink{}).Where("from_id = ? AND to_id = ? AND type = ?", ownerID, resourceID, OWNS).Count(&count)
	return count > 0
}

func (s *SharingStore) HasSecondaryOwnership(ownerID, resourceID string) bool {
	var count int64

	// Direct ownership (primary check)
	if s.HasOwnership(ownerID, resourceID) {
		return true
	}

	// Check for entities that the ownerID owns
	subQuery1 := s.db.Table("entity_links").Select("to_id").
		Where("from_id = ? AND type = ?", ownerID, OWNS)

	// Check for entities that own the ownerID
	subQuery2 := s.db.Table("entity_links").Select("from_id").
		Where("to_id = ? AND type = ?", ownerID, OWNS)

	// Check if any of the found entities own the resourceID
	s.db.Model(&EntityLink{}).
		Where("(from_id IN (?) OR from_id IN (?)) AND to_id = ? AND type = ?", subQuery1, subQuery2, resourceID, OWNS).
		Count(&count)

	return count > 0
}

func (s *SharingStore) ShareResource(resourceID uuid.UUID, fromOwnerID uuid.UUID, toEntityID uuid.UUID, roles string, permissions string, conditions string) error {
	if !s.HasOwnership(fromOwnerID.String(), resourceID.String()) {
		return fmt.Errorf("entity %s does not own resource %s", fromOwnerID, resourceID)
	}

	fromOwnerType := &Entity{}
	toEntityType := &Entity{}

	if er := s.db.Find(&fromOwnerType, "id = ?", toEntityID).Error; er != nil {
		fromOwnerType.Type = "unknown"
	}
	if er := s.db.Find(&toEntityType, "id = ?", resourceID).Error; er != nil {
		toEntityType.Type = "unknown"
	}

	shareEdge := &EntityLink{
		ID:          uuid.New(),
		SharedBy:    fromOwnerID.String(),
		FromID:      toEntityID,
		FromType:    fromOwnerType.Type,
		ToID:        resourceID,
		ToType:      toEntityType.Type,
		Type:        SHARES,
		Roles:       roles,
		Permissions: permissions,
		Conditions:  conditions,
	}

	return s.AddEdge(shareEdge)
}

func (s *SharingStore) RequestAccess(resourceOwner uuid.UUID, resourceID uuid.UUID, requesterID uuid.UUID, requestedRoles string) error {
	request := AccessRequest{
		ID:             uuid.New(),
		ResourceID:     resourceID,
		RequesterID:    requesterID,
		RequestedRoles: requestedRoles,
		ResourceOwner:  resourceOwner,
		Status:         "pending",
		Timestamp:      time.Now(),
	}
	return s.db.Create(&request).Error
}

func (s *SharingStore) GrantAccess(requestID uuid.UUID, granterID uuid.UUID, permissions string) error {
	var req AccessRequest
	if err := s.db.First(&req, requestID).Error; err != nil {
		return fmt.Errorf("access request not found")
	}

	if !s.HasOwnership(granterID.String(), req.ResourceID.String()) || !s.HasSecondaryOwnership(granterID.String(), req.ResourceID.String()) {
		return fmt.Errorf("entity %s cannot grant access to resource %s", granterID, req.ResourceID)
	}

	err := s.ShareResource(req.ResourceID, granterID, req.RequesterID, req.RequestedRoles, permissions, "")
	if err != nil {
		return err
	}

	req.Status = "granted"
	return s.db.Save(&req).Error
}
