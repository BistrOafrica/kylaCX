package service

import (
	"gorm.io/gorm"
)

type DepartmentStore struct {
	db *gorm.DB
}

func NewDepartmentStore(db *gorm.DB) *DepartmentStore {
	return &DepartmentStore{
		db: db,
	}
}

func (store *DepartmentStore) CreateDepartment(department *Department) (*Department, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(department).Error; err != nil {
			return err
		}
		if err := tx.Model(department).Association("Roles").Replace(department.Roles); err != nil {
			return err
		}
		if err := tx.Model(department).Association("Users").Replace(department.Users); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return department, nil
}

func (store *DepartmentStore) UpdateDepartment(department *Department) (*Department, error) {
	result := store.db.Omit("CreatedAt", "ID", "OwnerID", "OwnerType").Save(department).First(department)
	if result.Error != nil {
		return nil, result.Error
	}
	return department, nil
}

func (store *DepartmentStore) DeleteDepartment(departmentID string) error {
	result := store.db.Where("id = ?", departmentID).Delete(&Department{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (store *DepartmentStore) ReadDepartment(id string) (*Department, error) {
	var department Department
	result := store.db.Where("id = ?", id).First(&department)
	if result.Error != nil {
		return nil, result.Error
	}
	return &department, nil
}

func (store *DepartmentStore) ReadDepartments(scope *OpScope) ([]*Department, error) {
	var departments []*Department
	result := store.db.Where("owner_id = ? AND owner_type = ?", scope.ID, scope.Owner).Find(&departments)
	if result.Error != nil {
		return nil, result.Error
	}
	return departments, nil
}

func (store *DepartmentStore) ReadDepartmentsByBranch(branchId string, orgId string) ([]*Department, error) {
	var departments []*Department
	result := store.db.Where("branch_id = ? AND organisation_id = ?", branchId, orgId).Find(&departments)
	if result.Error != nil {
		return nil, result.Error
	}
	return departments, nil
}
