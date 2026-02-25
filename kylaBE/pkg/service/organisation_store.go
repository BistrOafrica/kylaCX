package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DatabaseOrganisationStore is a database implementation of the OrganisationStore
type OrganisationStore struct {
	db *gorm.DB
}

// NewOrganisationStore creates a new database organisation store
func NewOrganisationStore(db *gorm.DB) *OrganisationStore {
	return &OrganisationStore{
		db: db,
	}
}

func (store *OrganisationStore) Save(org *Organisation) (*Organisation, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {

		orgRoles := org.Roles
		org.Roles = nil
		orgUsers := org.Users
		org.Users = nil

		// Create organisation
		if err := tx.Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create organisation: %v", err)
		}
		// Add roles for organisation
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

// FindByID finds an organisation by ID
func (store *OrganisationStore) FindByID(id *uuid.UUID) (*Organisation, error) {
	var org Organisation
	if err := store.db.Preload(clause.Associations).Find(&org, "id = ?", id).Error; err != nil {
		log.Printf("failed to find organisation: %v", err)
		return nil, fmt.Errorf("failed to find organisation: %v", err)
	}
	return &org, nil
}

// FindByName finds an organisation by name
func (store *OrganisationStore) FindByName(name string) (*Organisation, error) {
	var org Organisation
	result := store.db.First(&org, "name = ?", name)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find organisation by name: %v", result.Error)
	}
	return &org, nil
}

// Update updates an organisation in the database
func (store *OrganisationStore) Update(org *Organisation) error {
	result := store.db.Omit(
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

// FindAll returns all organisations from the database
func (store *OrganisationStore) FindAll() ([]*Organisation, error) {
	var organisations []*Organisation
	result := store.db.Find(&organisations)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all organisations: %v", result.Error)
	}
	return organisations, nil
}

// Delete removes an organisation from the database
func (store *OrganisationStore) Delete(org *Organisation) error {
	result := store.db.Delete(org)
	if result.Error != nil {
		return fmt.Errorf("failed to delete organisation: %v", result.Error)
	}
	return nil
}

func (store *OrganisationStore) OrgWithTransaction(fn func(tx *gorm.DB) (*Organisation, error)) (*Organisation, error) {
	tx := store.db.Begin()

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
