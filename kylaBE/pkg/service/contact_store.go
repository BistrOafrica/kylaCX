package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DatabaseContactStore is a database implementation of the OrganisationStore
type ContactStore struct {
	db *gorm.DB
}

// NewContactStore creates a new database contact store
func NewContactStore(db *gorm.DB) *ContactStore {
	return &ContactStore{
		db: db,
	}
}

func (store *ContactStore) WithTransaction(fn func(tx *gorm.DB) error) error {
	tx := store.db.Begin()

	if tx.Error != nil {
		return errors.New("failed to start transaction")
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit transaction")
	}

	return nil
}

func (store *ContactStore) WithTransactionReturn(fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	tx := store.db.Begin()

	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}

	result, err := fn(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit transaction")
	}

	return result, nil
}

func (store *ContactStore) Save(contact *Contact) (*Contact, error) {
	fn := func(tx *gorm.DB) (*Contact, error) {
		tx.Begin()
		if result := tx.Create(contact); result.Error != nil {
			tx.Rollback()
			return nil, errors.New("failed to save contact")
		}
		return contact, nil
	}
	return fn(store.db)
}

func (store *ContactStore) SaveCustomField(customField *CustomFieldValue) (*CustomFieldValue, error) {
	fn := func(tx *gorm.DB) (*CustomFieldValue, error) {
		// Check if a record already exists for the given contact_id and custom_field_definition_id
		var existingCustomField CustomFieldValue
		err := tx.Where("contact_id = ? AND custom_field_definition_id = ?", customField.ContactID, customField.CustomFieldDefinitionID).
			First(&existingCustomField).Error

		if err == nil {
			// If the record exists, update it
			existingCustomField.Value = customField.Value
			if err := tx.Save(&existingCustomField).Error; err != nil {
				return nil, fmt.Errorf("failed to update custom field: %w", err)
			}
			return &existingCustomField, nil
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// If the record does not exist, insert it
			if err := tx.Create(customField).Error; err != nil {
				return nil, fmt.Errorf("failed to insert custom field: %w", err)
			}
			return customField, nil
		} else {
			// Handle other errors
			return nil, fmt.Errorf("failed to query custom field: %w", err)
		}
	}

	result, err := fn(store.db)
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (store *ContactStore) UpdateCustomField(customField *CustomFieldValue) (*CustomFieldValue, error) {
	fn := func(tx *gorm.DB) (*CustomFieldValue, error) {
		tx.Begin()
		if result := tx.Omit("ID", "ContactID", "CustomFieldDefinitionID", "OwnerID", "OwnerType").Save(customField); result.Error != nil {
			tx.Rollback()
			return nil, errors.New("failed to save custom field")
		}
		return customField, nil
	}
	result, err := fn(store.db)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindByName finds a contact by name
func (store *ContactStore) FindByName(name string) (*Contact, error) {
	var contact Contact
	result := store.db.Preload(clause.Associations).First(&contact, "first_name = ?", name) // Use "first_name" instead of "name"
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find contact by name: %v", result.Error)
	}
	return &contact, nil
}

// FindById finds a contact by ID
func (store *ContactStore) FindById(md *RequestMetadata, id string) (*Contact, error) {
	var contact Contact
	result := store.db.Preload(clause.Associations).First(&contact, "id = ?", id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find contact by ID: %v", result.Error)
	}

	opScope := &OpScope{
		Owner: OwnerType(contact.OwnerType),
		ID:    contact.OwnerID.String(),
	}

	if !CheckOpScope(md, opScope) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	return &contact, nil
}

func (store *ContactStore) FindByEmail(email string, scope *OpScope) (*Contact, error) {
	var contact Contact
	result := store.db.Preload(clause.Associations).First(&contact, "email = ? AND owner_id = ? AND owner_type = ?", email, scope.ID, scope.Owner)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find contact by email: %v", result.Error)
	}
	return &contact, nil
}

func (store *ContactStore) FindByPhone(phone string, scope *OpScope) (*Contact, error) {
	var contact Contact
	result := store.db.Preload(clause.Associations).First(&contact, "phone = ? AND owner_id = ? AND owner_type = ?", phone, scope.ID, scope.Owner)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find contact by phone: %v", result.Error)
	}
	return &contact, nil
}

// Update updates a contact in the database
func (store *ContactStore) Update(contact *Contact, scope *OpScope) (*Contact, error) {
	// Fields to exclude
	fieldsToExclude := []string{"Model", "ID", "CreatedAt", "DeletedAt", "OrganisationID", "SerialNumber", "CreatedBy", "OwnerID", "OwnerType"}
	// Read zero-valued fields
	fn := func(tx *gorm.DB) (*Contact, error) {
		tx.Begin()
		result := store.db.Omit(fieldsToExclude...).Save(contact).First(&contact, "id = ? AND owner_id = ? AND owner_type = ?", contact.ID, scope.ID, scope.Owner)
		if result.Error != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update contact: %v", result.Error)
		}
		return contact, nil
	}
	return fn(store.db)
}

// FindAll returns all contacts from the database associated with a specific organisation
func (store *ContactStore) FindAll(idsAllowingAccess []string, page int32, per_page int32) ([]*Contact, int32, error) {
	var contacts []*Contact
	var count int64
	offset := (page - 1) * per_page

	query := store.db.Where("owner_id IN (?)", idsAllowingAccess)
	result := query.Model(&Contact{}).Order("created_at desc")
	result.Count(&count)
	result = result.Preload(clause.Associations).Offset(int(offset)).Limit(int(per_page)).Find(&contacts)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find contacts: %v", result.Error)
	}
	return contacts, int32(count), nil
}

// Delete a contact from the database
func (store *ContactStore) Delete(id string, scope *OpScope) error {
	var contact Contact
	result := store.db.First(&contact, "id = ? AND owner_id = ? AND owner_type = ?", id, scope.ID, scope.Owner)
	if result.Error != nil {
		return fmt.Errorf("failed to find contact: %v", result.Error)
	}

	// Clear the associations in the join tables
	store.db.Model(&contact).Association("Tags").Clear()
	store.db.Model(&contact).Association("Labels").Clear()

	// Delete the custom field values associated with the contact
	store.db.Where("contact_id = ?", id).Delete(&CustomFieldValue{})

	// Delete the contact
	result = store.db.Delete(&contact)
	if result.Error != nil {
		return fmt.Errorf("failed to delete contact: %v", result.Error)
	}

	return nil
}

func (store *ContactStore) FindCustomFieldDefinitionByName(name string, scope *OpScope) (*CustomFieldDefinition, error) {
	var result *gorm.DB
	var definition *CustomFieldDefinition
	result = store.db.Where("name = ? AND owner_id = ? AND owner_type = ?", name, scope.ID, scope.Owner).First(&definition)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find custom field definition by name: %v", result.Error)
	}
	return definition, nil
}

func (store *ContactStore) SaveCustomFieldDefinition(customField *CustomFieldDefinition) (*CustomFieldDefinition, error) {
	fn := func(tx *gorm.DB) (*CustomFieldDefinition, error) {
		tx.Begin()
		var definition *CustomFieldDefinition
		if result := tx.Save(customField).Find(&definition); result.Error != nil {
			tx.Rollback()
			log.Println("failed to save custom field definition")
		}
		return definition, nil
	}
	return fn(store.db)
}

func (store *ContactStore) FindCustomFieldDefinitions(scope *OpScope) ([]*CustomFieldDefinition, error) {
	var customFieldDefinitions []*CustomFieldDefinition
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("owner_type = ? AND owner_id = ?", scope.Owner, scope.ID).Find(&customFieldDefinitions).Error
	})
	if err != nil {
		return nil, err
	}
	return customFieldDefinitions, nil
}

func (store *ContactStore) Search(organisationID uuid.UUID, branchID string, query string, page int32, perPage int32, scope *OpScope) ([]*Contact, int32, error) {
	var contacts []*Contact
	var count int64
	offset := (page - 1) * perPage
	result := store.db.Table("contacts").Joins("LEFT JOIN custom_field_values ON contacts.id = custom_field_values.contact_id").Joins("LEFT JOIN custom_field_definitions ON custom_field_values.custom_field_definition_id = custom_field_definitions.id AND custom_field_definitions.organisation_id = ?", organisationID).Where("contacts.organisation_id = ? AND contacts.branch_id = ? AND contacts.owner_id = ? AND contacts.owner_type = ? AND (LOWER(contacts.first_name) LIKE ? OR LOWER(contacts.last_name) LIKE ? OR LOWER(contacts.other_name) LIKE ? OR LOWER(contacts.nickname) LIKE ? OR LOWER(contacts.title) LIKE ? OR LOWER(contacts.prefix) LIKE ? OR LOWER(contacts.suffix) LIKE ? OR LOWER(contacts.email) LIKE ? OR LOWER(contacts.phone) LIKE ? OR LOWER(contacts.other_phone) LIKE ? OR LOWER(contacts.job_department) LIKE ? OR LOWER(contacts.job_title) LIKE ? OR LOWER(contacts.company) LIKE ? OR LOWER(contacts.notes) LIKE ? OR LOWER(contacts.country) LIKE ? OR LOWER(contacts.state) LIKE ? OR LOWER(contacts.city) LIKE ? OR LOWER(contacts.street) LIKE ? OR LOWER(contacts.postal_code) LIKE ? OR LOWER(contacts.url) LIKE ? OR LOWER(custom_field_values.value) LIKE ?)", organisationID, branchID, scope.ID, scope.Owner, "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%").Group("contacts.id")
	result.Count(&count)
	result = result.Preload(clause.Associations).Offset(int(offset)).Limit(int(perPage)).Find(&contacts)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to search contacts for organisation %s: %v", organisationID, result.Error)
	}
	return contacts, int32(count), nil
}

func (store *ContactStore) SearchWithinGroup(organisationID uuid.UUID, branchID string, query string, page int32, perPage int32, groupId string, scope *OpScope) ([]*Contact, int32, error) {
	var group ContactGroup
	result := store.db.Where("id = ? AND organisation_id = ?", groupId, organisationID).First(&group)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find group: %v", result.Error)
	}
	groupContactIds := make([]string, len(group.ContactIds))
	copy(groupContactIds, group.ContactIds)

	var contacts []*Contact
	var count int64
	offset := (page - 1) * perPage
	result = store.db.Table("contacts").Joins("LEFT JOIN custom_field_values ON contacts.id = custom_field_values.contact_id").Joins("LEFT JOIN custom_field_definitions ON custom_field_values.custom_field_definition_id = custom_field_definitions.id AND custom_field_definitions.organisation_id = ?", organisationID).Where("contacts.organisation_id = ? AND contacts.branch_id = ? AND contacts.owner_id = ? AND contacts.owner_type = ? AND contacts.id IN ? AND (LOWER(contacts.first_name) LIKE ? OR LOWER(contacts.last_name) LIKE ? OR LOWER(contacts.other_name) LIKE ? OR LOWER(contacts.nickname) LIKE ? OR LOWER(contacts.title) LIKE ? OR LOWER(contacts.prefix) LIKE ? OR LOWER(contacts.suffix) LIKE ? OR LOWER(contacts.email) LIKE ? OR LOWER(contacts.phone) LIKE ? OR LOWER(contacts.other_phone) LIKE ? OR LOWER(contacts.job_department) LIKE ? OR LOWER(contacts.job_title) LIKE ? OR LOWER(contacts.company) LIKE ? OR LOWER(contacts.notes) LIKE ? OR LOWER(contacts.country) LIKE ? OR LOWER(contacts.state) LIKE ? OR LOWER(contacts.city) LIKE ? OR LOWER(contacts.street) LIKE ? OR LOWER(contacts.postal_code) LIKE ? OR LOWER(contacts.url) LIKE ? OR LOWER(custom_field_values.value) LIKE ?)", organisationID, branchID, scope.ID, scope.Owner, groupContactIds, "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%").Group("contacts.id")
	result.Count(&count)
	result = result.Preload(clause.Associations).Offset(int(offset)).Limit(int(perPage)).Find(&contacts)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to search contacts for organisation %s: %v", organisationID, result.Error)
	}
	return contacts, int32(count), nil
}

func (store *ContactStore) SearchWithinAccount(organisationID uuid.UUID, branchID string, query string, page int32, perPage int32, accountId string, scope *OpScope) ([]*Contact, int32, error) {
	var contacts []*Contact
	var count int64
	offset := (page - 1) * perPage
	result := store.db.Table("contacts").Joins("LEFT JOIN custom_field_values ON contacts.id = custom_field_values.contact_id").Joins("LEFT JOIN custom_field_definitions ON custom_field_values.custom_field_definition_id = custom_field_definitions.id AND custom_field_definitions.organisation_id = ?", organisationID).Where("contacts.organisation_id = ? AND contacts.branch_id = ?  AND contacts.owner_id = ? AND contacts.owner_type = ? AND contacts.account_ids = ? AND (LOWER(contacts.first_name) LIKE ? OR LOWER(contacts.last_name) LIKE ? OR LOWER(contacts.other_name) LIKE ? OR LOWER(contacts.nickname) LIKE ? OR LOWER(contacts.title) LIKE ? OR LOWER(contacts.prefix) LIKE ? OR LOWER(contacts.suffix) LIKE ? OR LOWER(contacts.email) LIKE ? OR LOWER(contacts.phone) LIKE ? OR LOWER(contacts.other_phone) LIKE ? OR LOWER(contacts.job_department) LIKE ? OR LOWER(contacts.job_title) LIKE ? OR LOWER(contacts.company) LIKE ? OR LOWER(contacts.notes) LIKE ? OR LOWER(contacts.country) LIKE ? OR LOWER(contacts.state) LIKE ? OR LOWER(contacts.city) LIKE ? OR LOWER(contacts.street) LIKE ? OR LOWER(contacts.postal_code) LIKE ? OR LOWER(contacts.url) LIKE ? OR LOWER(custom_field_values.value) LIKE ?)", organisationID, branchID, scope.ID, scope.Owner, accountId, "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%").Group("contacts.id")
	result.Count(&count)
	result = result.Preload(clause.Associations).Offset(int(offset)).Limit(int(perPage)).Find(&contacts)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to search contacts for account %s: %v", accountId, result.Error)
	}
	return contacts, int32(count), nil
}

func (store *ContactStore) FindByEmailOrPhone(email string, phone string, scope *OpScope) (*Contact, error) {
	var contact Contact

	// If both email and phone are empty, return an empty contact
	if email == "" && phone == "" {
		return &contact, nil
	}

	queryParts := []string{"owner_id = ? AND owner_type = ?"}
	args := []interface{}{scope.ID, scope.Owner}

	if email != "" {
		queryParts = append(queryParts, "email = ?")
		args = append(args, email)
	}

	if phone != "" {
		queryParts = append(queryParts, "(phone = ? OR other_phone = ?)")
		args = append(args, phone, phone)
	}

	query := strings.Join(queryParts, " AND ")

	result := store.db.Where(query, args...).First(&contact)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			//return an empty contact object
			return &contact, nil
		} else {
			return nil, fmt.Errorf("failed to find contact by email or phone: %v", result.Error)
		}
	}
	return &contact, nil
}

func (store *ContactStore) FindContactChildren(contactId string, organisationID uuid.UUID) ([]*Contact, error) {
	var contacts []*Contact

	result := store.db.Where("parent_id = ? AND organisation_id = ?", contactId, organisationID).Preload(clause.Associations).Find(&contacts)

	if result.Error != nil {
		fmt.Printf("failed to find contact children: %v", result.Error)
	}

	return contacts, nil
}

func (store *ContactStore) FindAccountContacts(organisationID uuid.UUID, branchId string, page int32, per_page int32, accountId string) ([]*Contact, int32, error) {
	var contacts []*Contact
	var count int64
	offset := (page - 1) * per_page

	query := store.db.Where("organisation_id = ? AND branch_id = ? AND account_id = ?", organisationID, branchId, accountId)
	result := query.Model(&Contact{}).Order("created_at desc")
	result.Count(&count)
	result = result.Preload(clause.Associations).Offset(int(offset)).Limit(int(per_page)).Find(&contacts)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find contacts for account %s: %v", accountId, result.Error)
	}
	return contacts, int32(count), nil
}

func (store *ContactStore) FindCustomFieldByContactID(contactID string) ([]*CustomFieldValue, error) {
	var customFields []*CustomFieldValue
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("contact_id = ?", contactID).Find(&customFields).Error
	})
	if err != nil {
		return nil, err
	}
	return customFields, nil
}

func (store *ContactStore) FindCustomFieldDefinitionByID(id string) (*CustomFieldDefinition, error) {
	customFieldDefinition := &CustomFieldDefinition{}
	err := store.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).First(&customFieldDefinition)
		if result.Error != nil {
			return fmt.Errorf("failed to find custom field definition by ID: %v", result.Error)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return customFieldDefinition, nil
}

func (store *ContactStore) FindBySocialProfile(scope *OpScope, platformId string, pageId string, externalId string) (*Contact, error) {
	var contact Contact
	var socialProfile SocialProfile

	queryParts := []string{"platform_id = ?"}
	args := []interface{}{platformId}

	if pageId != "" {
		queryParts = append(queryParts, "page_id = ?")
		args = append(args, pageId)
	}

	if externalId != "" {
		queryParts = append(queryParts, "external_id = ?")
		args = append(args, externalId)
	}

	query := strings.Join(queryParts, " AND ")

	result := store.db.Table("social_profiles").Where(query, args...).First(&socialProfile)
	if result.Error == nil {
		result = store.db.Where("id = ?", socialProfile.ContactID).Preload(clause.Associations).First(&contact)
	}

	if contact.OwnerID != uuid.MustParse(scope.ID) && contact.OwnerType != scope.Owner {
		return nil, errors.New("failed to find contact by social profile")
	}

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			//return an empty contact object
			return &contact, nil
		} else {
			return nil, fmt.Errorf("failed to find contact by social profile: %v", result.Error)
		}
	}
	return &contact, nil
}

func (store *ContactStore) FindDuplicatesBasedOnEmailandPhone(ownerId string) (map[string][]*Contact, error) {
	// First, get all contacts with duplicate emails
	var emailDuplicates []*Contact
	res := store.db.Raw(`
		WITH duplicate_emails AS (
			SELECT email FROM contacts 
			WHERE owner_id = ? 
			AND email IS NOT NULL AND TRIM(email) != '' 
			GROUP BY email 
			HAVING COUNT(*) > 1
		)
		SELECT DISTINCT c.* FROM contacts c
		INNER JOIN duplicate_emails de ON c.email = de.email
		WHERE c.owner_id = ? AND c.deleted_at is NULL
	`, ownerId, ownerId).Find(&emailDuplicates)

	if res.Error != nil {
		return nil, res.Error
	}

	// Then, get all contacts with duplicate phone numbers (using normalized fields)
	var phoneDuplicates []*Contact
	res = store.db.Raw(`
		WITH duplicate_phones AS (
			SELECT phone_number FROM (
				SELECT DISTINCT id, normalized_phone as phone_number FROM contacts 
				WHERE owner_id = ? 
				AND normalized_phone IS NOT NULL AND TRIM(normalized_phone) != ''
				UNION
				SELECT DISTINCT id, normalized_other_phone as phone_number FROM contacts 
				WHERE owner_id = ? 
				AND normalized_other_phone IS NOT NULL AND TRIM(normalized_other_phone) != ''
			) all_phones
			GROUP BY phone_number
			HAVING COUNT(*) > 1
		)
		SELECT DISTINCT c.* FROM contacts c
		INNER JOIN duplicate_phones dp ON 
			c.normalized_phone = dp.phone_number OR 
			c.normalized_other_phone = dp.phone_number
		WHERE c.owner_id = ? AND c.deleted_at is NULL
	`, ownerId, ownerId, ownerId).Find(&phoneDuplicates)

	if res.Error != nil {
		return nil, res.Error
	}

	// Create a map to group duplicates
	duplicateGroups := make(map[string][]*Contact)

	// Group by email
	for _, contact := range emailDuplicates {
		if contact.Email != "" {
			key := "email:" + contact.Email
			duplicateGroups[key] = append(duplicateGroups[key], contact)

		}
	}

	// Group by normalized phone (both phone and other_phone)
	phoneGroups := make(map[string][]*Contact)
	for _, contact := range phoneDuplicates {
		if contact.NormalizedPhone != "" {
			key := "phone:" + contact.NormalizedPhone
			phoneGroups[key] = append(phoneGroups[key], contact)

		}
		if contact.NormalizedOtherPhone != "" {
			key := "phone:" + contact.NormalizedOtherPhone
			phoneGroups[key] = append(phoneGroups[key], contact)

		}
	}

	// Only add phone groups that have more than one contact
	for key, contacts := range phoneGroups {
		if len(contacts) > 1 {
			duplicateGroups[key] = contacts

		}
	}

	return duplicateGroups, nil
}

func (store *ContactStore) FindByIdsFromService(ids []string) ([]*Contact, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}

	var contacts []*Contact
	result := store.db.Preload(clause.Associations).Where("id IN (?)", ids).Find(&contacts)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find contacts by IDs: %v", result.Error)
	}

	return contacts, nil
}
