package branch

import (
	"errors"
	"fmt"
	"kyla-be/internal/authctx"

	"gorm.io/gorm"
)

// BranchStore manages persistence for Branch records.
type BranchStore struct {
	DB *gorm.DB
}

// NewBranchStore creates a new BranchStore.
func NewBranchStore(db *gorm.DB) *BranchStore {
	return &BranchStore{DB: db}
}

// Save persists a new branch inside a transaction.
func (bs *BranchStore) Save(branch *Branch) error {
	err := bs.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(branch).Error
	})
	if err != nil {
		return fmt.Errorf("failed to save branch: %v", err)
	}
	return nil
}

// FindByID retrieves a branch by its ID string.
func (bs *BranchStore) FindByID(id string) (*Branch, error) {
	var branch Branch
	result := bs.DB.Where("id = ?", id).First(&branch)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("branch not found with ID %s", id)
		}
		return nil, fmt.Errorf("failed to find branch: %v", result.Error)
	}
	return &branch, nil
}

// FindByOwner returns all branches belonging to the given scope.
func (bs *BranchStore) FindByOwner(scope *authctx.OpScope) ([]*Branch, error) {
	var branches []*Branch
	err := bs.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("owner_id = ? AND owner_type = ?", scope.ID, scope.Owner).Find(&branches).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find branches: %v", err)
	}
	return branches, nil
}

// FindDefaultBranch returns the default branch for the given scope.
func (bs *BranchStore) FindDefaultBranch(scope *authctx.OpScope) (*Branch, error) {
	var branch Branch
	err := bs.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("owner_id = ? AND owner_type = ? AND is_default = ?", scope.ID, scope.Owner, true).First(&branch).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find default branch: %v", err)
	}
	return &branch, nil
}

// Update saves changes to an existing branch, omitting immutable columns.
func (bs *BranchStore) Update(branch *Branch) error {
	err := bs.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Omit(
			"id",
			"created_at",
			"owner_id",
			"owner_type",
			"serial_number",
			"parent_id",
			"is_Default",
			"role_ids",
			"status",
			"OwnerID", "OwnerType",
		).Save(branch).Error
	})
	if err != nil {
		return fmt.Errorf("failed to update branch: %v", err)
	}
	return nil
}

// Delete removes a branch from the database by ID.
func (bs *BranchStore) Delete(id string) error {
	err := bs.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&Branch{}).Error; err != nil {
			return fmt.Errorf("failed to delete branch: %v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}
	return nil
}
