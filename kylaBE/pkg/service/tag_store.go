package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagStore struct {
	db *gorm.DB
}

func NewTagStore(db *gorm.DB) *TagStore {
	return &TagStore{
		db: db,
	}
}

func (store *TagStore) Save(tag *Tag) error {
	result := store.db.Create(tag)
	if result.Error != nil {
		return fmt.Errorf("failed to save tag: %v", result.Error)
	}
	return nil
}

func (store *TagStore) FindById(id string, organisationID uuid.UUID) (*Tag, error) {
	var tag Tag
	result := store.db.First(&tag, "id = ? AND organisation_id = ?", id, organisationID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tag not found with ID: %s", id)
		}
		return nil, fmt.Errorf("failed to find tag: %v", result.Error)
	}
	return &tag, nil
}

func (store *TagStore) Update(tag *Tag, id string, organisationID uuid.UUID) error {
	result := store.db.Model(&Tag{}).Omit("OrganisationID", "ID", "CreatedAt", "CreatedBy",
		"OwnerID", "OwnerType",
	).Where("id = ? AND organisation_id = ?", id, organisationID).Updates(tag)
	if result.Error != nil {
		return fmt.Errorf("failed to update tag: %v", result.Error)
	}
	return nil
}

func (store *TagStore) Delete(id string, organisationID uuid.UUID) error {
	result := store.db.Delete(&Tag{}, "id = ? AND organisation_id = ?", id, organisationID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete tag: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tag not found with ID: %s", id)
	}
	return nil
}

func (store *TagStore) FindByOrganisationID(organisationID uuid.UUID) ([]*Tag, error) {
	var tags []*Tag
	result := store.db.Find(&tags, "organisation_id = ?", organisationID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find tags: %v", result.Error)
	}
	return tags, nil
}
