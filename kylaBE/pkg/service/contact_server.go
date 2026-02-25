package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Contacts struct {
	pb.UnimplementedContactServiceServer
	contactStore *ContactStore
	AuthStore    *AuthStore
	BranchStore  *BranchStore
	GroupStore   *ContactGroupStore
	LeadClient   pb.LeadServiceClient
	SharingStore *SharingStore
}

func NewContactServer(
	contactStore *ContactStore,
	authStore *AuthStore,
	branchStore *BranchStore,
	groupStore *ContactGroupStore,
	LeadClient pb.LeadServiceClient,
	SharingStore *SharingStore,
) *Contacts {
	return &Contacts{
		contactStore: contactStore,
		AuthStore:    authStore,
		BranchStore:  branchStore,
		GroupStore:   groupStore,
		LeadClient:   LeadClient,
		SharingStore: SharingStore,
	}
}

func (c *Contacts) CreateContact(ctx context.Context, request *pb.CreateContactRequest) (*pb.CreateContactResponse, error) {
	log.Println("Create Contact")
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if !IsPbValidOwnerType(request.GetContact().GetOwnerType()) || request.GetContact().GetOwnerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid owner type or id")
	}
	scope := PbScopeToOpScope(&pb.Scope{
		OwnerType: request.GetContact().GetOwnerType(),
		OwnerId:   request.GetContact().GetOwnerId(),
	})
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	contact := PbContactToContact(request.GetContact())

	if err := utils.ValidateRequiredFields(k.NewConsts().ContactsRequiredFields, contact); err != nil {
		return nil, err
	}
	if !utils.ValidatePhoneFormat(contact.Phone) || !utils.ValidatePhoneFormat(contact.OtherPhone) {
		return nil, errors.New("invalid phone format")
	}

	contact.ID = uuid.New()
	contact.CreatedBy = contextData.UserID.String()
	contact.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["contacts"], contact.ID.String())

	contact.OwnerID = uuid.MustParse(scope.ID)
	contact.OwnerType = scope.Owner

	customFields := request.GetContact().GetCustomFields()
	var savedCustomFields []*CustomFieldValue
	var customFieldDefinitions []*CustomFieldDefinition
	var cfErr error

	// Save Contact
	newContact, err := c.contactStore.Save(contact)
	if err != nil {
		return nil, status.Error(500, "error while creating contact: %v")
	}
	// Save Custom Fields and new definitions
	if len(customFields) != 0 {
		//check if defined custom fields are existent, if not create them
		definitions, err := c.createContactFieldDefinitions(
			contextData,
			contact,
			customFields,
			scope,
			true,
		)
		if err != nil {
			log.Printf("error while creating custom field definitions: %v", err)
		}
		customFieldDefinitions = append(customFieldDefinitions, definitions...)

		savedCustomFields, cfErr = c.saveContactCustomFields(contact, customFields, definitions)
		if cfErr != nil {
			log.Printf("error while saving custom fields: %v", err)
		}
	}
	customFieldDefinitionNames := []string{}
	for _, definition := range customFieldDefinitions {
		customFieldDefinitionNames = append(customFieldDefinitionNames, definition.Name)
	}

	md := metadata.New(map[string]string{"authorization": contextData.Authorization})
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)
	log.Println("Sending contact to lead service", contextData.Authorization)

	if _, err := c.LeadClient.OnContactCreate(ctxWithToken, &pb.OnContactCreateRequest{
		Contact:                ContactToPbContact(newContact, customFieldDefinitions, savedCustomFields),
		CustomFieldDefinitions: customFieldDefinitionNames,
		Scope:                  &pb.Scope{OwnerType: pb.OwnerType(pb.OwnerType_value[string(scope.Owner)]), OwnerId: scope.ID},
	}); err != nil {
		log.Printf("error while sending contact to lead service: %v", err)
	}

	// TODO, create a node

	return &pb.CreateContactResponse{Contact: ContactToPbContact(newContact, customFieldDefinitions, savedCustomFields)}, nil
}

func (c *Contacts) ReadContact(ctx context.Context, request *pb.ReadContactRequest) (*pb.ReadContactResponse, error) {
	log.Println("Read Contact")
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	contact, err := c.contactStore.FindById(contextChanData, request.GetContactId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "contact not found")
	}
	fields, definitions, err := c.getContactCustomFields(contact.ID.String())
	if err != nil {
		contact.CustomFieldValues = []CustomFieldValue{}
	}
	return &pb.ReadContactResponse{
		Contact: ContactToPbContact(
			contact,
			definitions,
			fields,
		),
	}, nil
}

func (c *Contacts) ReadContacts(ctx context.Context, request *pb.ReadContactsRequest) (*pb.ReadContactsResponse, error) {
	log.Println("Read Contacts")
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}
	contextData, authErr := c.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read contacts")
	}

	contextDataForContacts, err := c.AuthStore.AuthInternalRequests(contextData.Authorization, "READ_CONTACTS")
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}

	idsAllowingAccess := GetScopeIDs(contextDataForContacts.Scopes)

	page := request.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := request.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	contacts, count, err := c.contactStore.FindAll(idsAllowingAccess, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("error while reading contacts: %v", err)
	}

	pbContacts := make([]*pb.Contact, 0, len(contacts))
	for _, contact := range contacts {
		fields, definitions, err := c.getContactCustomFields(contact.ID.String())
		if err != nil {
			contact.CustomFieldValues = []CustomFieldValue{}
		}
		pbContacts = append(pbContacts, ContactToPbContact(contact, definitions, fields))
	}

	return &pb.ReadContactsResponse{
		Contacts: pbContacts,
		Count:    count,
	}, nil
}

func (c *Contacts) UpdateContact(ctx context.Context, request *pb.UpdateContactRequest) (*pb.UpdateContactResponse, error) {
	log.Println("Update Contact")
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(&pb.Scope{
		OwnerType: request.GetContact().GetOwnerType(),
		OwnerId:   request.GetContact().GetOwnerId(),
	})
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	contact := PbContactToContact(request.GetContact())

	if err := utils.ValidateRequiredFields(k.NewConsts().ContactsRequiredFields, contact); err != nil {
		return nil, err
	}
	if !utils.ValidatePhoneFormat(contact.Phone) || !utils.ValidatePhoneFormat(contact.OtherPhone) {
		return nil, errors.New("invalid phone format")
	}

	customFields := request.GetContact().GetCustomFields()
	var savedCustomFields []*CustomFieldValue
	var customFieldDefinitions []*CustomFieldDefinition
	var cfErr error

	if len(customFields) != 0 {
		definitions, err := c.createContactFieldDefinitions(
			contextData,
			contact,
			customFields,
			scope,
			true,
		)
		if err != nil {
			log.Printf("error while creating custom field definitions: %v", err)
		}
		customFieldDefinitions = append(customFieldDefinitions, definitions...)
		savedCustomFields, cfErr = c.updateContactCustomFields(contact, customFields, definitions)
		if cfErr != nil {
			log.Printf("error while saving custom fields: %v", err)
		}
	}
	// Save Contact
	updatedContact, err := c.contactStore.Update(contact, scope)
	if err != nil {
		return nil, status.Error(500, "error while updating contact: %v")
	}

	return &pb.UpdateContactResponse{Contact: ContactToPbContact(updatedContact, customFieldDefinitions, savedCustomFields)}, nil
}

func (c *Contacts) DeleteContact(ctx context.Context, request *pb.DeleteContactRequest) (*pb.DeleteContactResponse, error) {
	log.Println("Delete Contact")
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())

	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	if err := c.GroupStore.RemoveContactFromGroups(request.GetContactId(), scope); err != nil {
		return nil, status.Error(codes.NotFound, "contact could not be removed from groups")
	}

	if err := c.contactStore.Delete(request.GetContactId(), scope); err != nil {
		return nil, status.Error(codes.NotFound, "contact not found")
	}

	return &pb.DeleteContactResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Contact deleted successfully",
		},
	}, nil
}

func (c *Contacts) ReadCustomFieldDefinitions(ctx context.Context, request *pb.ReadCustomFieldDefinitionsRequest) (*pb.ReadCustomFieldDefinitionsResponse, error) {
	log.Println("Update Contact")
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	customFieldDefs, err := c.contactStore.FindCustomFieldDefinitions(scope)

	if err != nil {
		return nil, fmt.Errorf("failed to read custom fields definitions: %v", err)
	}

	return &pb.ReadCustomFieldDefinitionsResponse{
		CustomFieldDefinitions: CustomFieldDefinitionsToPbCustomFieldDefinitions(customFieldDefs),
	}, nil
}

func (c *Contacts) BulkContactsImport(ctx context.Context, req *pb.BulkContactsImportRequest) (*pb.BulkContactsImportResponse, error) {
	log.Println("Bulk Contacts Import")

	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	// Initialize CSV reader
	r := csv.NewReader(strings.NewReader(string(req.Data)))

	// Read header row
	header, err := r.Read()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not read CSV header: %v", err)
	}

	// Create a map of header positions
	headerMap := make(map[string]int)
	for i, name := range header {
		headerMap[name] = i
	}

	// Handle contact group if provided
	group := PbContactGroupToContactGroup(req.GetGroup())

	// is a completely new group
	if group != nil && req.GetGroup().GetId() == "" {
		group.CreatedBy = contextData.UserID.String()
		group.ContactIds = []string{}
	}

	// Precompute field mappings
	standardFields := make(map[int]string)                     // column index → struct field name
	customFieldDefs := make(map[string]*CustomFieldDefinition) // column name → definition

	for i, column := range header {
		if utils.Includes(k.VALID_CONTACT_FIELDS(), column) {
			standardFields[i] = column
		} else {
			customColumn, err := c.createContactDefinition(&CustomFieldDefinition{
				Name:      column,
				Type:      "TEXT",
				OwnerType: scope.Owner,
				OwnerID:   uuid.MustParse(scope.ID),
			}, scope)
			if err != nil {
				log.Printf("error while creating custom field definition: %v", err)
				return nil, status.Errorf(codes.Unknown, "error while creating custom field definition: %v", err)
			}
			customFieldDefs[column] = customColumn
		}
	}

	// Process records
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error reading record: %v", err)
			continue
		}

		newContactId := uuid.New()
		contact := &Contact{
			ID:           newContactId,
			SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["contacts"], newContactId.String()),
			CreatedBy:    contextData.UserID.String(),
			OwnerID:      uuid.MustParse(scope.ID),
			OwnerType:    scope.Owner,
		}

		var customValues []*CustomFieldValue

		// Process standard fields
		for _, field := range k.VALID_CONTACT_FIELDS() {
			if pos, exists := headerMap[field]; exists && pos < len(record) && record[pos] != "" {
				switch field {
				case "first_name":
					contact.FirstName = record[pos]
				case "last_name":
					contact.LastName = record[pos]
				case "other_name":
					contact.OtherName = record[pos]
				case "phone":
					formattedPhone, valid := utils.FormatAndValidatePhone(record[pos])
					if valid {
						contact.Phone = formattedPhone
					}
				case "other_phone":
					formattedPhone, valid := utils.FormatAndValidatePhone(record[pos])
					if valid {
						contact.OtherPhone = formattedPhone
					}
				case "email":
					if utils.ValidateEmailFormat(record[pos]) {
						contact.Email = record[pos]
					}
				case "title":
					contact.Title = record[pos]
				case "prefix":
					contact.Prefix = record[pos]
				case "suffix":
					contact.Suffix = record[pos]
				case "birthday":
					if utils.ValidateDate(record[pos]) {
						contact.Birthday = record[pos]
					}
				case "url":
					contact.URL = record[pos]
				case "company":
					contact.Company = record[pos]
				case "job_title":
					contact.JobTitle = record[pos]
				case "job_department":
					contact.JobDepartment = record[pos]
				case "country":
					contact.Country = record[pos]
				case "state":
					contact.State = record[pos]
				case "city":
					contact.City = record[pos]
				case "street":
					contact.Street = record[pos]
				case "postal_code":
					contact.PostalCode = record[pos]
				case "notes":
					contact.Notes = record[pos]
				case "nickname":
					contact.Nickname = record[pos]
				case "is_vip":
					contact.IsVip = record[pos] == "true" || record[pos] == "1"
				}
			}
		}

		// Process custom fields
		for colName, pos := range headerMap {
			if !utils.Includes(k.VALID_CONTACT_FIELDS(), colName) && pos < len(record) && record[pos] != "" {
				if customDef, exists := customFieldDefs[colName]; exists {
					customValues = append(customValues, &CustomFieldValue{
						ID:                      uuid.New(),
						ContactID:               contact.ID,
						CustomFieldDefinitionID: customDef.ID.String(),
						Value:                   record[pos],
					})
				}
			}
		}

		// Save contact
		newContact, err := c.contactStore.Save(contact)
		if err != nil {
			log.Printf("error while saving contact: %v", err)
			continue
		}

		// Save custom fields
		for _, value := range customValues {
			value.ContactID = newContact.ID
			if _, err := c.contactStore.SaveCustomField(value); err != nil {
				log.Printf("error while saving custom field value: %v", err)
				continue
			}
		}

		// Update group if specified
		if group != nil {
			contact.GroupIds = []string{group.ID.String()}
			group.ContactIds = append(group.ContactIds, contact.ID.String())
		}
	}

	// Save group if specified
	if group != nil {
		if err := c.GroupStore.Save(group); err != nil {
			log.Printf("error while saving group: %v", err)
			return nil, status.Errorf(codes.Unknown, "error while saving group: %v", err)
		}
	}

	return &pb.BulkContactsImportResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Contacts uploaded successfully",
		},
	}, nil
}

func (c *Contacts) SearchContacts(ctx context.Context, request *pb.SearchContactsRequest) (*pb.SearchContactsResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	page := request.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := request.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	query := request.GetQuery()

	groupId := request.GetGroupId()

	accountId := request.GetAccountId()

	var contacts []*Contact
	var count int32
	var err error

	if groupId != "" {
		contacts, count, err = c.contactStore.SearchWithinGroup(contextData.OrganisationID, contextData.BranchID.String(), query, page, perPage, groupId, scope)

		if err != nil {
			return nil, fmt.Errorf("error while searching contacts: %v", err)
		}

		return &pb.SearchContactsResponse{
			Contacts: c.ContactsToPbContacts(contacts),
			Count:    count,
		}, nil
	} else if accountId != "" {
		contacts, count, err = c.contactStore.SearchWithinAccount(contextData.OrganisationID, contextData.BranchID.String(), query, page, perPage, accountId, scope)

		if err != nil {
			return nil, fmt.Errorf("error while searching contacts: %v", err)
		}

		return &pb.SearchContactsResponse{
			Contacts: c.ContactsToPbContacts(contacts),
			Count:    count,
		}, nil
	} else {
		contacts, count, err = c.contactStore.Search(contextData.OrganisationID, contextData.BranchID.String(), query, page, perPage, scope)
		if err != nil {
			return nil, fmt.Errorf("error while searching contacts: %v", err)
		}

		return &pb.SearchContactsResponse{
			Contacts: c.ContactsToPbContacts(contacts),
			Count:    count,
		}, nil
	}
}

func (c *Contacts) FindContactByEmailOrPhone(_ context.Context, request *pb.FindContactByEmailOrPhoneRequest) (*pb.FindContactByEmailOrPhoneResponse, error) {
	email, phone := request.GetEmail(), request.GetPhone()
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())
	contact, err := c.contactStore.FindByEmailOrPhone(email, phone, scope)
	if err != nil {
		return nil, status.Error(codes.NotFound, "contact not found")
	}

	fields, definitions, err := c.getContactCustomFields(contact.ID.String())
	if err != nil {
		contact.CustomFieldValues = []CustomFieldValue{}
	}

	return &pb.FindContactByEmailOrPhoneResponse{
		Contact: ContactToPbContact(
			contact,
			definitions,
			fields,
		),
	}, nil
}

func (c *Contacts) CreateContactWithoutToken(_ context.Context, request *pb.CreateContactWithoutTokenRequest) (*pb.CreateContactWithoutTokenResponse, error) {
	log.Println("Internal Service Create Contact")
	contact := PbContactToContact(request.GetContact())
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())
	if err := utils.ValidateRequiredFields(k.NewConsts().ContactsRequiredFields, contact); err != nil {
		return nil, err
	}
	if !utils.ValidatePhoneFormat(contact.Phone) || !utils.ValidatePhoneFormat(contact.OtherPhone) {
		return nil, errors.New("invalid phone format")
	}
	contact.ID = uuid.New()
	contact.CreatedBy = "USERS"
	contact.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["contacts"], contact.ID.String())
	contact.OwnerType = scope.Owner
	contact.OwnerID = uuid.MustParse(scope.ID)
	customFields := request.GetContact().GetCustomFields()
	var savedCustomFields []*CustomFieldValue
	var customFieldDefinitions []*CustomFieldDefinition
	var cfErr error
	newContact, err := c.contactStore.Save(contact)
	if err != nil {
		return nil, status.Error(500, "error while creating contact: %v")
	}
	// Save Custom Fields and new definitions
	if len(customFields) != 0 {
		//check if defined custom fields are existent, if not create them
		scopes := &Scopes{}
		if scope.ID != "" {
			switch scope.Owner {
			case OwnerType(ORGANISATIONS):
				scopes.Organisation = uuid.MustParse(scope.ID)
			case OwnerType(BRANCHES):
				scopes.Branch = uuid.MustParse(scope.ID)
			case OwnerType(TEAMS):
				scopes.Teams = []string{scope.ID}
			case OwnerType(DEPARTMENTS):
				scopes.Departments = []string{scope.ID}
			case OwnerType(USERS):
				scopes.User = uuid.MustParse(scope.ID)
			default:
				scopes.Organisation = uuid.MustParse(scope.ID)
			}
		}
		definitions, err := c.createContactFieldDefinitions(
			&RequestMetadata{
				Scopes: scopes,
			},
			contact,
			customFields,
			scope,
			false,
		)
		if err != nil {
			log.Printf("error while creating custom field definitions: %v", err)
		}
		customFieldDefinitions = append(customFieldDefinitions, definitions...)

		savedCustomFields, cfErr = c.saveContactCustomFields(contact, customFields, definitions)
		if cfErr != nil {
			log.Printf("error while saving custom fields: %v", err)
		}
	}

	return &pb.CreateContactWithoutTokenResponse{Contact: ContactToPbContact(newContact, customFieldDefinitions, savedCustomFields)}, nil
}

func (c *Contacts) ReadContactChildren(ctx context.Context, request *pb.ReadContactChildrenRequest) (*pb.ReadContactChildrenResponse, error) {
	log.Println("Read Contact children")
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}

	contacts, err := c.contactStore.FindContactChildren(request.GetContactId(), contextData.OrganisationID)
	if err != nil {
		return nil, fmt.Errorf("error while searching contacts: %v", err)
	}

	pbContacts := make([]*pb.Contact, len(contacts))
	for _, contact := range contacts {
		fields, definitions, err := c.getContactCustomFields(contact.ID.String())
		if err != nil {
			contact.CustomFieldValues = []CustomFieldValue{}
		}
		pbContacts = append(pbContacts, ContactToPbContact(contact, definitions, fields))
	}

	return &pb.ReadContactChildrenResponse{
		Contacts: pbContacts,
	}, nil
}

func (c *Contacts) ReadAccountContacts(ctx context.Context, request *pb.ReadAccountContactsRequest) (*pb.ReadAccountContactsResponse, error) {
	log.Println("Read Account Contacts")
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())
	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}
	if request.GetAccountId() == "" {
		return nil, status.Error(500, "Account ID is required")
	}

	page := request.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := request.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	contacts, count, err := c.contactStore.FindAccountContacts(contextData.OrganisationID, contextData.BranchID.String(), page, perPage, request.GetAccountId())
	if err != nil {
		return nil, fmt.Errorf("error while reading contacts: %v", err)
	}

	pbContacts := make([]*pb.Contact, len(contacts))
	for _, contact := range contacts {
		fields, definitions, err := c.getContactCustomFields(contact.ID.String())
		if err != nil {
			contact.CustomFieldValues = []CustomFieldValue{}
		}
		pbContacts = append(pbContacts, ContactToPbContact(contact, definitions, fields))
	}

	return &pb.ReadAccountContactsResponse{
		Contacts: pbContacts,
		Count:    count,
	}, nil
}

func (c *Contacts) createContactFieldDefinitions(contextData *RequestMetadata, contact *Contact, customFields []*pb.CustomField, scope *OpScope, withToken bool) ([]*CustomFieldDefinition, error) {
	customFieldDefs := make([]*CustomFieldDefinition, 0, len(customFields))
	for _, customField := range customFields {
		if customField.Name != "" {
			customFieldDef, err := c.contactStore.FindCustomFieldDefinitionByName(customField.Name, scope)
			if err != nil {
				customFieldDefId := uuid.New()
				customFieldDef = &CustomFieldDefinition{
					ID:        customFieldDefId,
					Name:      customField.Name,
					Type:      customField.Type,
					OwnerType: scope.Owner,
					OwnerID:   uuid.MustParse(scope.ID),
				}
				if withToken {
					definitionScope, err := c.AuthStore.InternalScopeCheck(contextData, "ADD_OWNER_TO_CUSTOM_FIELD_DEFINITION")
					if err != nil {
						log.Printf("Error while adding contact owner to custom field definition")
						break
					}
					if definitionScope.RequestAuth == k.NewConsts().TRUE {
						customFieldDef.OwnerType = contact.OwnerType
						customFieldDef.OwnerID = uuid.MustParse(contact.OwnerID.String())
					}
				}
				customFieldDefinition, err := c.contactStore.SaveCustomFieldDefinition(customFieldDef)
				if err != nil {
					log.Printf("Error while saving custom field definition")
				} else {
					customFieldDefs = append(customFieldDefs, customFieldDefinition)
				}
			} else {
				customFieldDefs = append(customFieldDefs, customFieldDef)
			}
		}
	}
	return customFieldDefs, nil
}

func (c *Contacts) saveContactCustomFields(contact *Contact, customFields []*pb.CustomField, customFieldDefinitions []*CustomFieldDefinition) ([]*CustomFieldValue, error) {
	savedCustomFields := make([]*CustomFieldValue, 0)

	for _, definition := range customFieldDefinitions {
		for _, customField := range customFields {
			if customField.Name == definition.Name {
				// Create or update the custom field value
				newCustomFieldValue := ContactCustomFieldValuesMaker(contact.ID, definition.ID.String(), customField.Value)

				// Save the custom field value
				savedCustomField, err := c.contactStore.SaveCustomField(newCustomFieldValue)
				if err != nil {
					log.Printf("error while saving custom field value: %v", err)
					return nil, fmt.Errorf("failed to save custom field: %w", err)
				}

				// Append the saved custom field to the result
				savedCustomFields = append(savedCustomFields, savedCustomField)
			}
		}
	}

	return savedCustomFields, nil
}

// update contact custom fields
func (c *Contacts) updateContactCustomFields(contact *Contact, customFields []*pb.CustomField, definitions []*CustomFieldDefinition) ([]*CustomFieldValue, error) {
	savedCustomFields := make([]*CustomFieldValue, 0, len(customFields))
	for _, customField := range customFields {
		for _, definition := range definitions {
			if customField.ValueId == "" {
				customField.ValueId = uuid.New().String()
			}
			if customField.DefinitionID == "" && customField.Name == definition.Name {
				customField.DefinitionID = definition.ID.String()
			}
			customFieldValueID, err := uuid.Parse(customField.ValueId)
			if err != nil {
				log.Printf("error while parsing custom field value ID: %v", err)
			}
			customFields, err := c.contactStore.UpdateCustomField(&CustomFieldValue{
				ID:                      customFieldValueID,
				ContactID:               contact.ID,
				CustomFieldDefinitionID: customField.DefinitionID,
				Value:                   customField.Value,
			})
			if err != nil {
				log.Printf("error while saving custom field value: %v", err)
			}
			savedCustomFields = append(savedCustomFields, customFields)
		}
	}
	return savedCustomFields, nil
}

func (c *Contacts) getContactCustomFields(contactId string) (fields []*CustomFieldValue, definitions []*CustomFieldDefinition, err error) {
	customFields, err := c.contactStore.FindCustomFieldByContactID(contactId)
	if err != nil {
		return nil, nil, err
	}
	var defList []*CustomFieldDefinition
	for _, field := range customFields {
		definition, err := c.contactStore.FindCustomFieldDefinitionByID(field.CustomFieldDefinitionID)
		if err != nil {
			return nil, nil, err
		}
		defList = append(defList, definition)
	}
	return customFields, defList, nil
}

func (c *Contacts) createContactDefinition(definition *CustomFieldDefinition, scope *OpScope) (*CustomFieldDefinition, error) {
	customFieldDef, err := c.contactStore.FindCustomFieldDefinitionByName(definition.Name, scope)
	if err != nil {
		customFieldDefId := uuid.New()
		customFieldDef = &CustomFieldDefinition{
			ID:        customFieldDefId,
			Name:      definition.Name,
			Type:      definition.Type,
			OwnerType: scope.Owner,
			OwnerID:   uuid.MustParse(scope.ID),
		}
		customFieldDef, err = c.contactStore.SaveCustomFieldDefinition(customFieldDef)
		if err != nil {
			return nil, status.Error(500, "error while creating custom field: %v")
		}
	}
	return customFieldDef, nil
}

func (c *Contacts) ContactsToPbContacts(contacts []*Contact) []*pb.Contact {
	var pbContacts []*pb.Contact
	for _, contact := range contacts {
		fields, definitions, err := c.getContactCustomFields(contact.ID.String())
		if err != nil {
			log.Printf("Error getting custom fields for contact %s: %v", contact.ID, err)
			// Continue with empty custom fields rather than failing
			fields = []*CustomFieldValue{}
			definitions = []*CustomFieldDefinition{}
		}

		pbContact := ContactToPbContact(contact, definitions, fields)
		if pbContact != nil {
			pbContacts = append(pbContacts, pbContact)
		} else {
			log.Printf("Failed to convert contact %s to protobuf", contact.ID)
		}
	}
	return pbContacts
}

func (c *Contacts) FindContactBySocialProfile(_ context.Context, request *pb.FindContactBySocialProfileRequest) (*pb.FindContactBySocialProfileResponse, error) {
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}
	scope := PbScopeToOpScope(request.GetScope())

	if request.GetPlatformId() == "" {
		return nil, errors.New("platform id required")
	}

	if request.GetExternalUserId() == "" {
		return nil, errors.New("external id required")
	}
	contact, err := c.contactStore.FindBySocialProfile(scope, request.GetPlatformId(), request.GetPageId(), request.GetExternalUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "contact not found")
	}
	fields, definitions, err := c.getContactCustomFields(contact.ID.String())
	if err != nil {
		contact.CustomFieldValues = []CustomFieldValue{}
	}

	return &pb.FindContactBySocialProfileResponse{
		Contact: ContactToPbContact(
			contact,
			definitions,
			fields,
		),
	}, nil
}

func (c *Contacts) FindContactDuplicates(ctx context.Context, request *pb.FindContactDuplicatesRequest) (*pb.FindContactDuplicatesResponse, error) {
	log.Println("Find Contact Duplicates")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}
	contextData, authErr := c.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to find duplicate contacts")
	}

	idsAllowingAccess := GetScopeIDs(contextData.Scopes)

	// Create a map to track contacts we've already seen
	seenContacts := make(map[string]bool)
	duplicateGroups := make(map[string][]*Contact)

	for _, id := range idsAllowingAccess {
		duplicateGroupsForID, err := c.contactStore.FindDuplicatesBasedOnEmailandPhone(id)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "error while finding contact duplicates: %v", err)
		}

		// Process each group
		for key, contacts := range duplicateGroupsForID {
			// Skip groups with only one contact
			if len(contacts) <= 1 {
				continue
			}

			// Initialize the group if it doesn't exist
			if _, exists := duplicateGroups[key]; !exists {
				duplicateGroups[key] = make([]*Contact, 0)
			}

			// Add contacts that haven't been seen before
			for _, contact := range contacts {
				if !seenContacts[contact.ID.String()] {
					duplicateGroups[key] = append(duplicateGroups[key], contact)
					seenContacts[contact.ID.String()] = true
				}
			}
		}
	}

	// Filter out groups that ended up with only one contact after deduplication
	finalGroups := make(map[string][]*Contact)
	for key, contacts := range duplicateGroups {
		if len(contacts) > 1 {
			finalGroups[key] = contacts
		}
	}

	var pbGroups []*pb.DuplicateGroup
	for key, contacts := range finalGroups {
		pbContacts := c.ContactsToPbContacts(contacts)

		// Extract the type (email/phone) and value from the key
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			log.Printf("Invalid key format: %s", key)
			continue
		}

		pbGroups = append(pbGroups, &pb.DuplicateGroup{
			Type:     parts[0],
			Value:    parts[1],
			Contacts: pbContacts,
		})
	}

	return &pb.FindContactDuplicatesResponse{
		DuplicateGroups: pbGroups,
	}, nil
}

func (c *Contacts) MergeDuplicateContacts(ctx context.Context, request *pb.MergeDuplicateContactsRequest) (*pb.MergeDuplicateContactsResponse, error) {
	log.Println("Merge Duplicate Contacts")

	if len(request.GetContactIds()) < 2 {
		return nil, status.Error(codes.InvalidArgument, "At least 2 contacts are required")
	}

	if request.GetSelectedMasterContactId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Selected master contact ID is required")
	}

	// Check permissions
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	// Find the master contact (the one to keep)
	var masterContact *Contact
	var otherContacts []*Contact

	contacts := make([]*Contact, 0)
	for _, contactId := range request.GetContactIds() {
		contact, err := c.contactStore.FindById(contextData, contactId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "contact not found: %v", err)
		}

		if contact.ID.String() == request.GetSelectedMasterContactId() {
			masterContact = contact
		} else {
			otherContacts = append(otherContacts, contact)
		}
		contacts = append(contacts, contact)
	}

	// Validate all contacts have the same owner
	firstOwnerID := contacts[0].OwnerID.String()
	firstOwnerType := contacts[0].OwnerType

	for _, contact := range contacts {
		if contact.OwnerID.String() != firstOwnerID || contact.OwnerType != firstOwnerType {
			return nil, status.Error(codes.InvalidArgument, "All contacts must have the same owner")
		}
	}

	if !CheckIfIDInScope(contextData.Scopes, firstOwnerID) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	// Start database transaction
	tx := c.contactStore.db.Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to start transaction: %v", tx.Error)
	}

	if masterContact == nil {
		tx.Rollback()
		return nil, status.Error(codes.InvalidArgument, "selected master contact not found in the list")
	}

	// Merge data from other contacts into master contact
	for _, contact := range otherContacts {
		// Merge basic fields (keep non-empty values from other contacts)
		if masterContact.FirstName == "" && contact.FirstName != "" {
			masterContact.FirstName = contact.FirstName
		}
		if masterContact.LastName == "" && contact.LastName != "" {
			masterContact.LastName = contact.LastName
		}
		if masterContact.OtherName == "" && contact.OtherName != "" {
			masterContact.OtherName = contact.OtherName
		}
		if masterContact.Title == "" && contact.Title != "" {
			masterContact.Title = contact.Title
		}
		if masterContact.Prefix == "" && contact.Prefix != "" {
			masterContact.Prefix = contact.Prefix
		}
		if masterContact.Suffix == "" && contact.Suffix != "" {
			masterContact.Suffix = contact.Suffix
		}
		if masterContact.JobDepartment == "" && contact.JobDepartment != "" {
			masterContact.JobDepartment = contact.JobDepartment
		}
		if masterContact.JobTitle == "" && contact.JobTitle != "" {
			masterContact.JobTitle = contact.JobTitle
		}
		if masterContact.Company == "" && contact.Company != "" {
			masterContact.Company = contact.Company
		}
		if masterContact.Nickname == "" && contact.Nickname != "" {
			masterContact.Nickname = contact.Nickname
		}
		if masterContact.Notes == "" && contact.Notes != "" {
			masterContact.Notes = contact.Notes
		}
		if masterContact.Birthday == "" && contact.Birthday != "" {
			masterContact.Birthday = contact.Birthday
		}
		if masterContact.URL == "" && contact.URL != "" {
			masterContact.URL = contact.URL
		}
		if masterContact.Country == "" && contact.Country != "" {
			masterContact.Country = contact.Country
		}
		if masterContact.State == "" && contact.State != "" {
			masterContact.State = contact.State
		}
		if masterContact.City == "" && contact.City != "" {
			masterContact.City = contact.City
		}
		if masterContact.Street == "" && contact.Street != "" {
			masterContact.Street = contact.Street
		}
		if masterContact.PostalCode == "" && contact.PostalCode != "" {
			masterContact.PostalCode = contact.PostalCode
		}
		if masterContact.Email == "" && contact.Email != "" {
			masterContact.Email = contact.Email
		}

		// Merge phone numbers (keep unique numbers)
		if masterContact.Phone == "" && contact.Phone != "" {
			masterContact.Phone = contact.Phone
			masterContact.NormalizedPhone = contact.NormalizedPhone
		}
		if masterContact.OtherPhone == "" && contact.OtherPhone != "" {
			masterContact.OtherPhone = contact.OtherPhone
			masterContact.NormalizedOtherPhone = contact.NormalizedOtherPhone
		}

		// Merge arrays (append unique values)
		if len(contact.GroupIds) > 0 {
			existingGroups := make(map[string]bool)
			for _, id := range masterContact.GroupIds {
				existingGroups[id] = true
			}
			for _, id := range contact.GroupIds {
				if !existingGroups[id] {
					masterContact.GroupIds = append(masterContact.GroupIds, id)
				}
			}
		}

		if len(contact.PipelineIDs) > 0 {
			existingPipelines := make(map[string]bool)
			for _, id := range masterContact.PipelineIDs {
				existingPipelines[id] = true
			}
			for _, id := range contact.PipelineIDs {
				if !existingPipelines[id] {
					masterContact.PipelineIDs = append(masterContact.PipelineIDs, id)
				}
			}
		}

		if len(contact.StageIDs) > 0 {
			existingStages := make(map[string]bool)
			for _, id := range masterContact.StageIDs {
				existingStages[id] = true
			}
			for _, id := range contact.StageIDs {
				if !existingStages[id] {
					masterContact.StageIDs = append(masterContact.StageIDs, id)
				}
			}
		}

		// Merge social profiles
		if len(contact.SocialProfiles) > 0 {
			// Update contact ID for all social profiles
			for _, profile := range contact.SocialProfiles {
				profile.ContactID = masterContact.ID
				if err := tx.Save(profile).Error; err != nil {
					tx.Rollback()
					return nil, status.Errorf(codes.Internal, "failed to update social profile: %v", err)
				}
			}
		}

		// Merge tags (many-to-many relationship)
		if len(contact.Tags) > 0 {
			existingTags := make(map[uuid.UUID]bool)
			for _, tag := range masterContact.Tags {
				existingTags[tag.ID] = true
			}
			for _, tag := range contact.Tags {
				if !existingTags[tag.ID] {
					if err := tx.Model(masterContact).Association("Tags").Append(tag); err != nil {
						tx.Rollback()
						return nil, status.Errorf(codes.Internal, "failed to add tag: %v", err)
					}
				}
			}
		}

		// Merge labels (many-to-many relationship)
		if len(contact.Labels) > 0 {
			existingLabels := make(map[uuid.UUID]bool)
			for _, label := range masterContact.Labels {
				existingLabels[label.ID] = true
			}
			for _, label := range contact.Labels {
				if !existingLabels[label.ID] {
					if err := tx.Model(masterContact).Association("Labels").Append(label); err != nil {
						tx.Rollback()
						return nil, status.Errorf(codes.Internal, "failed to add label: %v", err)
					}
				}
			}
		}

		// Merge custom fields
		otherFields, _, err := c.getContactCustomFields(contact.ID.String())
		if err == nil {
			masterFields, _, _ := c.getContactCustomFields(masterContact.ID.String())

			// Create a map of existing custom fields in master contact
			masterFieldMap := make(map[string]*CustomFieldValue)
			for _, field := range masterFields {
				masterFieldMap[field.CustomFieldDefinitionID] = field
			}

			// Add or update custom fields from other contact
			for _, field := range otherFields {
				if existingField, exists := masterFieldMap[field.CustomFieldDefinitionID]; exists {
					// Update if empty in master
					if existingField.Value == "" && field.Value != "" {
						existingField.Value = field.Value
						if err := tx.Save(existingField).Error; err != nil {
							tx.Rollback()
							return nil, status.Errorf(codes.Internal, "failed to update custom field: %v", err)
						}
					}
				} else {
					// Add new custom field
					newField := &CustomFieldValue{
						ID:                      uuid.New(),
						ContactID:               masterContact.ID,
						CustomFieldDefinitionID: field.CustomFieldDefinitionID,
						Value:                   field.Value,
					}
					if err := tx.Create(newField).Error; err != nil {
						tx.Rollback()
						return nil, status.Errorf(codes.Internal, "failed to create custom field: %v", err)
					}
				}
			}

			// Delete custom field values of the merged contact
			if err := tx.Where("contact_id = ?", contact.ID).Delete(&CustomFieldValue{}).Error; err != nil {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "failed to delete merged contact's custom fields: %v", err)
			}
		}

		// Create merge history record
		mergeHistory := &ContactMergeHistory{
			ID:              uuid.New(),
			MasterContactID: masterContact.ID,
			MergedContactID: contact.ID,
			MergedAt:        time.Now(),
			MergedBy:        contextData.UserID.String(),
		}
		if err := tx.Create(mergeHistory).Error; err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "failed to create merge history: %v", err)
		}

		// Soft delete the merged contact
		if err := tx.Model(&Contact{}).Where("id = ?", contact.ID).Update("deleted_at", time.Now()).Error; err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "failed to delete merged contact: %v", err)
		}
	}

	// Update the master contact
	if err := tx.Save(masterContact).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "failed to update master contact: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &pb.MergeDuplicateContactsResponse{
		Success: true,
	}, nil
}

func (c *Contacts) BulkContactsExport(request *pb.BulkContactsExportRequest, stream pb.ContactService_BulkContactsExportServer) error {
	log.Println("Bulk Contacts Export")

	// Get context from the stream
	ctx := stream.Context()

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go c.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	contextData, authErr := c.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to bulk export contacts")
	}

	contextDataForContacts, err := c.AuthStore.AuthInternalRequests(contextData.Authorization, "BULK_CONTACTS_EXPORT")
	if err != nil {
		log.Println("Error: ", err)
		return err
	}

	idsAllowingAccess := GetScopeIDs(contextDataForContacts.Scopes)

	var allFetchedContacts []*Contact

	if request.GetGroupId() != "" {
		group, err := c.GroupStore.FindByID(request.GetGroupId(), contextDataForContacts)
		if err != nil {
			return status.Errorf(codes.Unknown, "error while fetching group %v", err)
		}

		if group != nil {
			contacts, count, err := c.GroupStore.ReadGroupContacts(group.ID.String(), 1, 100)

			if err != nil {
				return status.Errorf(codes.Unknown, "error while fetching contacts %v", err)
			}

			allFetchedContacts = append(allFetchedContacts, contacts...)

			if count > 100 {
				noOfFetches := (count + 99) / 100 // Ensure rounding up

				for i := int32(1); i < noOfFetches; i++ {
					moreContacts, _, err := c.GroupStore.ReadGroupContacts(group.ID.String(), i*100+1, 100)
					if err != nil {
						return status.Errorf(codes.Unknown, "error while fetching contacts %v", err)
					}

					allFetchedContacts = append(allFetchedContacts, moreContacts...)
				}
			}
		}
	} else {
		contacts, count, err := c.contactStore.FindAll(idsAllowingAccess, 1, 100)

		if err != nil {
			return status.Errorf(codes.Unknown, "error while fetching contacts %v", err)
		}

		allFetchedContacts = append(allFetchedContacts, contacts...)

		if count > 100 {
			noOfFetches := (count + 99) / 100 // Ensure rounding up

			for i := int32(1); i < noOfFetches; i++ {
				moreContacts, _, err := c.contactStore.FindAll(idsAllowingAccess, i*100+1, 100)
				if err != nil {
					return status.Errorf(codes.Unknown, "error while fetching contacts %v", err)
				}

				allFetchedContacts = append(allFetchedContacts, moreContacts...)
			}
		}
	}

	var formattedScopes []*OpScope

	formattedScopes = append(formattedScopes, &OpScope{
		Owner: OwnerType(USERS),
		ID:    contextDataForContacts.Scopes.User.String(),
	})
	for _, team := range contextDataForContacts.Scopes.Teams {
		formattedScopes = append(formattedScopes, &OpScope{
			Owner: OwnerType(TEAMS),
			ID:    team,
		})
	}
	for _, department := range contextDataForContacts.Scopes.Departments {
		formattedScopes = append(formattedScopes, &OpScope{
			Owner: OwnerType(DEPARTMENTS),
			ID:    department,
		})
	}
	formattedScopes = append(formattedScopes, &OpScope{
		Owner: OwnerType(BRANCHES),
		ID:    contextDataForContacts.Scopes.Branch.String(),
	})
	formattedScopes = append(formattedScopes, &OpScope{
		Owner: OwnerType(ORGANISATIONS),
		ID:    contextDataForContacts.Scopes.Organisation.String(),
	})

	var allDefs []*CustomFieldDefinition
	for _, scope := range formattedScopes {
		customFieldDefs, err := c.contactStore.FindCustomFieldDefinitions(scope)
		if err != nil {
			log.Println(codes.Unknown, "error while fetching custom fields %v", err)
		}
		allDefs = append(allDefs, customFieldDefs...)
	}

	// Create a map for quick lookup of custom field definitions by ID
	defsMap := make(map[string]string)
	for _, def := range allDefs {
		defsMap[def.ID.String()] = def.Name
	}

	// Prepare header row
	headers := []string{
		"title", "first_name", "last_name", "phone", "other_phone", "email",
		"prefix", "suffix", "other_name", "nickname", "birthday", "company",
		"job_title", "job_department", "url", "country", "state", "city",
		"street", "postal_code", "notes", "is_vip",
	}

	// Add custom field headers
	for _, def := range allDefs {
		headers = append(headers, def.Name)
	}

	// Create a buffer for CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	writer.Write(headers)
	writer.Flush()

	// Send header chunk
	if err := stream.Send(&pb.BulkContactsExportResponse{
		ChunkData: buf.Bytes(),
	}); err != nil {
		return err
	}

	// Process each contact in chunks
	buf.Reset()
	const chunkSize = 100 // Number of contacts per chunk
	var currentChunk [][]string

	for i, contact := range allFetchedContacts {
		record := []string{
			contact.Title, contact.FirstName, contact.LastName, "\t" + contact.Phone,
			"\t" + contact.OtherPhone, contact.Email, contact.Prefix, contact.Suffix,
			contact.OtherName, contact.Nickname, contact.Birthday, contact.Company,
			contact.JobTitle, contact.JobDepartment, contact.URL, contact.Country,
			contact.State, contact.City, contact.Street, contact.PostalCode,
			contact.Notes, strconv.FormatBool(contact.IsVip),
		}

		// Add custom field values
		for _, def := range allDefs {
			value := ""
			for _, cf := range contact.CustomFieldValues {
				if cf.CustomFieldDefinitionID == def.ID.String() {
					value = cf.Value
					break
				}
			}
			record = append(record, value)
		}

		currentChunk = append(currentChunk, record)

		// Send chunk when we reach chunkSize or at the end
		if len(currentChunk) >= chunkSize || i == len(allFetchedContacts)-1 {
			buf.Reset()
			writer = csv.NewWriter(&buf)
			for _, r := range currentChunk {
				writer.Write(r)
			}
			writer.Flush()

			if err := stream.Send(&pb.BulkContactsExportResponse{
				ChunkData: buf.Bytes(),
			}); err != nil {
				return err
			}

			currentChunk = currentChunk[:0] // Reset chunk
		}
	}

	return nil
}

func (c *Contacts) ReadByIdsFromService(ctx context.Context, request *pb.ReadContactsByIdsFromServiceRequest) (*pb.ReadContactsByIdsFromServiceResponse, error) {
	log.Println("Read Contacts from Internal Service")

	contacts, err := c.contactStore.FindByIdsFromService(request.GetContactIds())
	if err != nil {
		return nil, err
	}

	pbContacts := make([]*pb.Contact, 0, len(contacts))
	for _, contact := range contacts {
		fields, definitions, err := c.getContactCustomFields(contact.ID.String())
		if err != nil {
			contact.CustomFieldValues = []CustomFieldValue{}
		}
		pbContacts = append(pbContacts, ContactToPbContact(contact, definitions, fields))
	}

	return &pb.ReadContactsByIdsFromServiceResponse{
		Contacts: pbContacts,
	}, nil
}
