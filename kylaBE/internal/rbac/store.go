package rbac

import (
	"fmt"
	"kyla-be/internal/authctx"
	"kyla-be/pkg/utils"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RbacStore handles persistence for Role records.
type RbacStore struct {
	DB *gorm.DB
}

// NewRbacStore creates a new RbacStore.
func NewRbacStore(db *gorm.DB) *RbacStore {
	return &RbacStore{
		DB: db,
	}
}

// SaveRole creates or updates a Role.
func (store *RbacStore) SaveRole(role *Role) (*Role, error) {
	result := store.DB.Save(role).Find(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role, nil
}

// FindRoleByID retrieves a Role by its UUID string.
func (store *RbacStore) FindRoleByID(id string) (*Role, error) {
	var role Role
	result := store.DB.First(&role, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

// FindByName retrieves a Role matching name and scope.
func (store *RbacStore) FindByName(name string, scope *authctx.OpScope) (*Role, error) {
	var role Role
	result := store.DB.First(&role, "name = ? AND owner_type = ? AND owner_id = ?", name, scope.Owner, scope.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

// FindRolesByIDs retrieves multiple roles by UUID slice.
func (store *RbacStore) FindRolesByIDs(ids []uuid.UUID) ([]*Role, error) {
	var roles []*Role
	result := store.DB.Where("id IN ?", ids).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

// FindDefaultRole retrieves the default (basic) role for an organisation.
func (store *RbacStore) FindDefaultRole(organisationID uuid.UUID) (*Role, error) {
	var role Role
	result := store.DB.First(&role, "owner_type = ? AND owner_id = ? AND is_default = ?", authctx.ORGANISATIONS, organisationID, true)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

// DeleteRole soft-deletes a role by ID string.
func (store *RbacStore) DeleteRole(roleID string) error {
	result := store.DB.Where("id = ?", roleID).Delete(&Role{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindAll retrieves all roles.
func (store *RbacStore) FindAll() ([]*Role, error) {
	var roles []*Role
	result := store.DB.Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

// FindAllByScope retrieves all roles for a given ownership scope.
func (store *RbacStore) FindAllByScope(scope *authctx.OpScope) ([]*Role, error) {
	var roles []*Role
	result := store.DB.Find(&roles, "owner_type = ? AND owner_id = ?", scope.Owner, scope.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

// UpdateRole saves changes to a Role, omitting immutable fields.
func (store *RbacStore) UpdateRole(role *Role) error {
	result := store.DB.Omit(
		"OwnerID", "OwnerType", "CreatedAt",
		"ID", "SerialNumber",
	).Save(role)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindRolesByPermission retrieves roles that include the given permission within a scope.
func (store *RbacStore) FindRolesByPermission(permission string, scope *authctx.OpScope) ([]*Role, error) {
	var roles []*Role
	result := store.DB.Where("? = ANY(permission_code_names) AND owner_type = ? AND owner_id = ?", permission, scope.Owner, scope.ID).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

// AddPermissionToRole appends permission code names to a role.
func (store *RbacStore) AddPermissionToRole(roleID string, permissionCodeNames []string) (*Role, error) {
	role, err := store.FindRoleByID(roleID)
	if err != nil {
		return nil, err
	}
	role.PermissionCodeNames = append(role.PermissionCodeNames, permissionCodeNames...)
	return role, store.UpdateRole(role)
}

// RemovePermissionFromRole removes permission code names from a role.
func (store *RbacStore) RemovePermissionFromRole(roleID string, permissionCodeNames []string) (*Role, error) {
	role, err := store.FindRoleByID(roleID)
	if err != nil {
		return nil, err
	}
	role.PermissionCodeNames = utils.RemoveStringsFromSlice(role.PermissionCodeNames, permissionCodeNames)
	return role, store.UpdateRole(role)
}

// NameExtractor defines how to extract a display name from any model instance.
type NameExtractor func(interface{}) string

// IdNameMapping is a lightweight ID-to-Name pair used in scope resolution.
type IdNameMapping = authctx.IdNameMapping

// FindDynamicIdNameMapping queries any GORM model by ID and extracts its name.
func (store *RbacStore) FindDynamicIdNameMapping(
	model interface{},
	id string,
	nameExtractor NameExtractor,
) (IdNameMapping, error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	newModel := reflect.New(modelType).Interface()

	result := store.DB.Where("id = ?", id).First(newModel)
	if result.Error != nil {
		return IdNameMapping{}, result.Error
	}

	val := reflect.ValueOf(newModel).Elem()
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
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
