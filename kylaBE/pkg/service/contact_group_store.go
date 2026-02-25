package service

import (
	"fmt"
	"kyla-be/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DatabaseContactGroupStore is an implementation of ContactGroupStore using a database.
type ContactGroupStore struct {
	db *gorm.DB
}

func NewContactGroupStore(db *gorm.DB) *ContactGroupStore {
	return &ContactGroupStore{db: db}
}

func (store *ContactGroupStore) Save(group *ContactGroup) error {

	result := store.db.Create(group)
	if result.Error != nil {
		return fmt.Errorf("failed to save contact group: %v", result.Error)
	}
	return nil
}

func (store *ContactGroupStore) FindByID(id string, md *RequestMetadata) (*ContactGroup, error) {
	var group ContactGroup
	result := store.db.Where("id = ?", id).First(&group)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find contact group: %v", result.Error)
	}

	opScope := &OpScope{
		Owner: OwnerType(group.OwnerType),
		ID:    group.OwnerID.String(),
	}

	if !CheckOpScope(md, opScope) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	return &group, nil
}

func (store *ContactGroupStore) FindAll(idsAllowingAccess []string, page int32, per_page int32) ([]*ContactGroup, int32, error) {
	var groups []*ContactGroup
	var count int64
	offset := (page - 1) * per_page
	query := store.db.Where("owner_id IN (?)", idsAllowingAccess)

	result := query.Model(&ContactGroup{}).Order("created_at desc")
	result.Count(&count)
	result = result.Offset(int(offset)).Limit(int(per_page)).Find(&groups)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find contact groups: %v", result.Error)
	}
	return groups, int32(count), nil
}

func (store *ContactGroupStore) Update(group *ContactGroup, scope *OpScope) (*ContactGroup, error) {
	// Fields to exclude
	fieldsToExclude := []string{"Model", "ID", "CreatedAt", "DeletedAt", "SerialNumber", "CreatedBy", "OwnerID", "OwnerType"}

	// Read zero-valued fields
	zeroFields := utils.FilterObjectWithZeroValues(group, fieldsToExclude)

	result := store.db.Model(&ContactGroup{}).Omit(fieldsToExclude...).Where("id = ? AND owner_type = ? AND owner_id = ?", group.ID, scope.Owner, scope.ID).Updates(group)

	if len(zeroFields) > 0 {
		result = result.UpdateColumns(zeroFields)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update group: %v", result.Error)
	}

	if len(group.ContactIds) == 0 {
		group.ContactIds = nil
		result = store.db.Model(&ContactGroup{}).Where("id = ? AND owner_type = ? owner_id = ?", group.ID, scope.Owner, scope.ID).UpdateColumn("ContactIds", group.ContactIds)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update contact group: %v", result.Error)
	}
	return group, nil
}

func (store *ContactGroupStore) Delete(id string, scope *OpScope) error {
	result := store.db.Where("id = ? AND owner_type = ? AND owner_id = ?", id, scope.Owner, scope.ID).Delete(&ContactGroup{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete contact group: %v", result.Error)
	}
	return nil
}

func (store *ContactGroupStore) ReadGroupContacts(groupId string, page int32, perPage int32) ([]*Contact, int32, error) {
	var contacts []*Contact
	// retrieve contact ids from the group
	var group *ContactGroup
	var count int64
	offset := (page - 1) * perPage

	result := store.db.Where("id = ?", groupId).First(&group)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find group: %v", result.Error)
	}

	contactIds := []string(group.ContactIds)
	// get contacts from the contact ids in group.ContactIds
	secondQuery := store.db.Where("id IN (?)", contactIds)
	result = secondQuery.Model(&Contact{}).Order("created_at desc")
	result.Count(&count)
	result = secondQuery.Preload(clause.Associations).Offset(int(offset)).Limit(int(perPage)).Find(&contacts)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find contacts for group: %v", result.Error)
	}

	return contacts, int32(count), nil
}

func (store *ContactGroupStore) RemoveContactFromGroups(contactId string, scope *OpScope) error {
	result := store.db.Model(&ContactGroup{}).Where("contact_ids @> ARRAY[?]::text[] AND owner_type = ? AND owner_id = ?", contactId, scope.Owner, scope.ID).Update("contact_ids", gorm.Expr("array_remove(contact_ids, ?)", contactId))
	if result.Error != nil {
		return fmt.Errorf("failed to remove contact from groups: %v", result.Error)
	}
	return nil
}
