package label

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LabelStore struct {
	db *gorm.DB
}

func NewLabelStore(db *gorm.DB) *LabelStore {
	return &LabelStore{
		db: db,
	}
}

func (store *LabelStore) Save(label *Label) error {
	result := store.db.Create(label)
	if result.Error != nil {
		return fmt.Errorf("failed to save label: %v", result.Error)
	}
	return nil
}

func (store *LabelStore) FindById(id string, organisationID uuid.UUID) (*Label, error) {
	var label Label
	result := store.db.First(&label, "id = ? AND organisation_id = ?", id, organisationID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("label not found with ID: %s", id)
		}
		return nil, fmt.Errorf("failed to find label: %v", result.Error)
	}
	return &label, nil
}

func (store *LabelStore) Update(label *Label, id string, organisationID uuid.UUID) error {
	result := store.db.Model(&Label{}).Omit("OrganisationID", "ID", "CreatedAt", "CreatedBy").Where("id = ? AND organisation_id = ?", id, organisationID).Updates(label)
	if result.Error != nil {
		return fmt.Errorf("failed to update label: %v", result.Error)
	}
	return nil
}

func (store *LabelStore) Delete(id string, organisationID uuid.UUID) error {
	result := store.db.Delete(&Label{}, "id = ? AND organisation_id = ?", id, organisationID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete label: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("label not found with ID: %s", id)
	}
	return nil
}

func (store *LabelStore) FindByOrganisationID(organisationID uuid.UUID) ([]*Label, error) {
	var labels []*Label
	result := store.db.Find(&labels, "organisation_id = ?", organisationID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find labels: %v", result.Error)
	}
	return labels, nil
}
