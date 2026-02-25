package service

import (
	"fmt"
	"kyla-be/pkg/utils"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RbacStore struct {
	db *gorm.DB
}

func NewRbacStore(db *gorm.DB) *RbacStore {
	return &RbacStore{
		db: db,
	}
}

// Roles
func (store *RbacStore) SaveRole(role *Role) (*Role, error) {
	result := store.db.Save(role).Find(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role, nil
}

func (store *RbacStore) FindRoleByID(id string) (*Role, error) {
	var role Role
	result := store.db.First(&role, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func (store *RbacStore) FindByName(name string, scope *OpScope) (*Role, error) {
	var role Role
	result := store.db.First(&role, "name = ? AND owner_type = ? AND owner_id = ?", name, scope.Owner, scope.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func (store *RbacStore) FindRolesByIDs(ids []uuid.UUID) ([]*Role, error) {
	var roles []*Role
	result := store.db.Where("id IN ?", ids).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func (store *RbacStore) FindDefaultRole(organisationID uuid.UUID) (*Role, error) {
	var role Role
	result := store.db.First(&role, "owner_type = ? AND owner_id = ? AND is_default = ?", ORGANISATIONS, organisationID, true)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func (store *RbacStore) DeleteRole(roleID string) error {
	result := store.db.Where("id = ?", roleID).Delete(&Role{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (store *RbacStore) FindAll() ([]*Role, error) {
	var roles []*Role
	result := store.db.Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func (store *RbacStore) FindAllByScope(scope *OpScope) ([]*Role, error) {
	var roles []*Role
	result := store.db.Find(&roles, "owner_type = ? AND owner_id = ?", scope.Owner, scope.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func (store *RbacStore) UpdateRole(role *Role) error {
	result := store.db.Omit(
		"OwnerID", "OwnerType", "CreatedAt",
		"ID", "SerialNumber",
	).Save(role)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (store *RbacStore) FindRolesByPermission(permission string, scope *OpScope) ([]*Role, error) {
	var roles []*Role
	result := store.db.Where("? = ANY(permission_code_names) AND owner_type = ? AND owner_id = ?", permission, scope.Owner, scope.ID).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func (store *RbacStore) AddPermissionToRole(roleID string, permissionCodeNames []string) (*Role, error) {
	role, err := store.FindRoleByID(roleID)
	if err != nil {
		return nil, err
	}
	role.PermissionCodeNames = append(role.PermissionCodeNames, permissionCodeNames...)
	return role, store.UpdateRole(role)
}

func (store *RbacStore) RemovePermissionFromRole(roleID string, permissionCodeNames []string) (*Role, error) {
	role, err := store.FindRoleByID(roleID)
	if err != nil {
		return nil, err
	}
	role.PermissionCodeNames = utils.RemoveStringsFromSlice(role.PermissionCodeNames, permissionCodeNames)
	return role, store.UpdateRole(role)
}

func (store *RbacStore) FindDynamicIdNameMapping(
	model interface{},
	id string,
	nameExtractor NameExtractor,
) (IdNameMapping, error) {
	// Create new instance of model
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	newModel := reflect.New(modelType).Interface()

	// Execute query using GORM's default primary key lookup
	result := store.db.Where("id = ?", id).First(newModel)
	if result.Error != nil {
		return IdNameMapping{}, result.Error
	}

	// Use reflection to find the ID field automatically
	val := reflect.ValueOf(newModel).Elem()
	idField := val.FieldByName("ID") // Try default GORM convention
	if !idField.IsValid() {
		// Fallback to "Id" if "ID" not found
		idField = val.FieldByName("Id")
	}
	if !idField.IsValid() {
		return IdNameMapping{}, fmt.Errorf("model doesn't have standard ID field")
	}

	return IdNameMapping{
		ID:   fmt.Sprintf("%v", idField.Interface()),
		Name: nameExtractor(newModel),
	}, nil
}
