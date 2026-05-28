package organisation

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrganisationStore handles persistence for Organisation records.
type OrganisationStore struct {
	DB *gorm.DB
}

// NewOrganisationStore creates a new OrganisationStore.
func NewOrganisationStore(db *gorm.DB) *OrganisationStore {
	return &OrganisationStore{
		DB: db,
	}
}

// Save persists a new Organisation together with its roles and user associations.
func (store *OrganisationStore) Save(org *Organisation) (*Organisation, error) {
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		orgRoles := org.Roles
		org.Roles = nil
		orgUsers := org.Users
		org.Users = nil

		if err := tx.Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create organisation: %v", err)
		}
		for _, role := range orgRoles {
			if err := tx.Create(&role).Error; err != nil {
				log.Printf("failed to create role: %v", err)
				return fmt.Errorf("failed to create role: %v", err)
			}
		}
		if err := tx.Model(&org).Association("Users").Replace(orgUsers); err != nil {
			return fmt.Errorf("failed to add user to organisation: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save organisation: %v", err)
	}
	return org, nil
}

// FindByID retrieves an Organisation by UUID, preloading associations.
func (store *OrganisationStore) FindByID(id *uuid.UUID) (*Organisation, error) {
	var org Organisation
	if err := store.DB.Preload(clause.Associations).Find(&org, "id = ?", id).Error; err != nil {
		log.Printf("failed to find organisation: %v", err)
		return nil, fmt.Errorf("failed to find organisation: %v", err)
	}
	return &org, nil
}

// FindByName retrieves an Organisation by name.
func (store *OrganisationStore) FindByName(name string) (*Organisation, error) {
	var org Organisation
	result := store.DB.First(&org, "name = ?", name)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find organisation by name: %v", result.Error)
	}
	return &org, nil
}

// Update saves changes to an Organisation, omitting immutable fields.
func (store *OrganisationStore) Update(org *Organisation) error {
	result := store.DB.Omit(
		"id",
		"created_at",
		"updated_at",
		"deleted_at",
		"serial_number",
		"referral_code",
		"status",
		"owner_type",
		"owner_id",
	).Save(org)
	if result.Error != nil {
		return fmt.Errorf("failed to update organisation: %v", result.Error)
	}
	return nil
}

// FindAll returns all organisations.
func (store *OrganisationStore) FindAll() ([]*Organisation, error) {
	var organisations []*Organisation
	result := store.DB.Find(&organisations)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all organisations: %v", result.Error)
	}
	return organisations, nil
}

// Delete soft-deletes an organisation.
func (store *OrganisationStore) Delete(org *Organisation) error {
	result := store.DB.Delete(org)
	if result.Error != nil {
		return fmt.Errorf("failed to delete organisation: %v", result.Error)
	}
	return nil
}

// OrgWithTransaction runs fn inside a transaction, rolling back on error.
func (store *OrganisationStore) OrgWithTransaction(fn func(tx *gorm.DB) (*Organisation, error)) (*Organisation, error) {
	tx := store.DB.Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}
	org, err := fn(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit transaction")
	}
	return org, nil
}
