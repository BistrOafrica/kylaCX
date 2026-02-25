package service

import (
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"strings"

	"github.com/google/uuid"
)

func PbContactToContact(contact *pb.Contact) *Contact {
	tagIds := contact.GetTagIds()
	tags := make([]Tag, len(tagIds))
	for i, id := range tagIds {
		tagID, err := uuid.Parse(id)
		if err != nil {
			tagID = uuid.New()
		}
		tags[i] = Tag{ID: tagID}
	}

	labelIds := contact.GetLabelIds()
	labels := make([]Label, len(labelIds))
	for i, id := range labelIds {
		labelID, err := uuid.Parse(id)
		if err != nil {
			labelID = uuid.New()
		}
		labels[i] = Label{ID: labelID}
	}

	id, err := uuid.Parse(contact.GetId())
	if err != nil {
		id = uuid.New()
	}

	parentID, err := uuid.Parse(contact.GetParentId())
	if err != nil {
		parentID = uuid.Nil
	}

	socialProfiles := PbSocialProfilesToSocialProfiles(contact.GetSocialProfiles())
	var finalSocialProfiles []SocialProfile
	for _, p := range socialProfiles {
		finalSocialProfiles = append(finalSocialProfiles, *p)
	}

	return &Contact{
		ID:             id,
		SerialNumber:   contact.GetSerialNumber(),
		FirstName:      contact.GetFirstName(),
		LastName:       contact.GetLastName(),
		OtherName:      contact.GetOtherName(),
		Nickname:       contact.GetNickname(),
		Email:          contact.GetEmail(),
		Phone:          contact.GetPhone(),
		OtherPhone:     contact.GetOtherPhone(),
		CreatedBy:      contact.GetCreatedBy(),
		Title:          contact.GetTitle(),
		Prefix:         contact.GetPrefix(),
		Suffix:         contact.GetSuffix(),
		JobDepartment:  contact.GetJobDepartment(),
		JobTitle:       contact.GetJobTitle(),
		Company:        contact.GetCompany(),
		Notes:          contact.GetNotes(),
		Birthday:       contact.GetBirthday(),
		URL:            contact.GetUrl(),
		Country:        contact.GetCountry(),
		State:          contact.GetState(),
		City:           contact.GetCity(),
		Street:         contact.GetStreet(),
		PostalCode:     contact.GetPostalCode(),
		Tags:           tags,
		Labels:         labels,
		OwnerID:        uuid.MustParse(contact.GetOwnerId()),
		OwnerType:      OwnerType(pb.OwnerType_name[int32(contact.GetOwnerType())]),
		ParentId:       parentID,
		SocialProfiles: finalSocialProfiles,
		IsVip:          contact.GetIsVip(),
	}
}

func ContactToPbContact(contact *Contact, definitions []*CustomFieldDefinition, customFields []*CustomFieldValue) *pb.Contact {
	tagIds := make([]string, len(contact.Tags))
	for i, tag := range contact.Tags {
		tagIds[i] = tag.ID.String()
	}

	tags := make([]*pb.Tag, len(contact.Tags))
	for i, tag := range contact.Tags {
		tags[i] = &pb.Tag{
			Id:        tag.ID.String(),
			Name:      tag.Name,
			ColorCode: tag.ColorCode,
			CreatedBy: tag.CreatedBy,
			CreatedAt: tag.CreatedAt.String(),
			UpdatedAt: tag.UpdatedAt.String(),
			OwnerType: pb.OwnerType(pb.OwnerType_value[string(tag.OwnerType)]),
			OwnerId:   tag.OwnerID.String(),
		}
	}

	labelIds := make([]string, len(contact.Labels))
	for i, label := range contact.Labels {
		labelIds[i] = label.ID.String()
	}

	labels := make([]*pb.Label, len(contact.Labels))
	for i, label := range contact.Labels {
		labels[i] = &pb.Label{
			Id:        label.ID.String(),
			Name:      label.Name,
			CreatedBy: label.CreatedBy,
			CreatedAt: label.CreatedAt.String(),
			UpdatedAt: label.UpdatedAt.String(),
			OwnerType: pb.OwnerType(pb.OwnerType_value[string(label.OwnerType)]),
			OwnerId:   label.OwnerID.String(),
		}
	}

	CustomFields := make([]*pb.CustomField, 0, len(customFields))
	for _, field := range customFields {
		for _, definition := range definitions {
			if field.CustomFieldDefinitionID == definition.ID.String() {
				customFieldValue := &pb.CustomField{
					Name:         definition.Name,
					Value:        field.Value,
					Type:         definition.Type,
					ValueId:      field.ID.String(),
					DefinitionID: definition.ID.String(),
				}
				CustomFields = append(CustomFields, customFieldValue)
			}
		}
	}

	var socialProfiles []*SocialProfile
	for _, socialProfile := range contact.SocialProfiles {
		socialProfiles = append(socialProfiles, &socialProfile)
	}

	return &pb.Contact{
		Id:             contact.ID.String(),
		SerialNumber:   contact.SerialNumber,
		FirstName:      contact.FirstName,
		LastName:       contact.LastName,
		OtherName:      contact.OtherName,
		Nickname:       contact.Nickname,
		Email:          contact.Email,
		Phone:          contact.Phone,
		OtherPhone:     contact.OtherPhone,
		TagIds:         tagIds,
		LabelIds:       labelIds,
		Tags:           tags,
		Labels:         labels,
		CreatedBy:      contact.CreatedBy,
		Title:          contact.Title,
		Prefix:         contact.Prefix,
		Suffix:         contact.Suffix,
		JobDepartment:  contact.JobDepartment,
		JobTitle:       contact.JobTitle,
		Company:        contact.Company,
		Notes:          contact.Notes,
		Birthday:       contact.Birthday,
		Url:            contact.URL,
		Country:        contact.Country,
		State:          contact.State,
		City:           contact.City,
		Street:         contact.Street,
		PostalCode:     contact.PostalCode,
		CreatedAt:      contact.CreatedAt.String(),
		UpdatedAt:      contact.UpdatedAt.String(),
		CustomFields:   CustomFields,
		OwnerId:        contact.OwnerID.String(),
		OwnerType:      pb.OwnerType(pb.OwnerType_value[string(contact.OwnerType)]),
		ParentId:       contact.ParentId.String(),
		SocialProfiles: SocialProfilesToPbSocialProfiles(socialProfiles),
		IsVip:          contact.IsVip,
	}
}

func PbContactsToContacts(contacts []*pb.Contact) []*Contact {
	var cs []*Contact
	for _, contact := range contacts {
		cs = append(cs, PbContactToContact(contact))
	}
	return cs
}

func PbTagToTag(tag *pb.Tag) *Tag {
	id, err := uuid.Parse(tag.GetId())
	if err != nil {
		id = uuid.New()
	}
	return &Tag{
		ID:        id,
		ColorCode: tag.GetColorCode(),
		Name:      tag.GetName(),
		CreatedBy: tag.GetCreatedBy(),
		OwnerType: OwnerType(tag.GetOwnerType()),
		OwnerID:   uuid.MustParse(tag.GetOwnerId()),
	}
}

func TagToPbTag(tag *Tag) *pb.Tag {
	return &pb.Tag{
		Id:        tag.ID.String(),
		ColorCode: tag.ColorCode,
		Name:      tag.Name,
		CreatedBy: tag.CreatedBy,
		CreatedAt: tag.CreatedAt.String(),
		UpdatedAt: tag.UpdatedAt.String(),
		OwnerType: pb.OwnerType(pb.OwnerType_value[string(tag.OwnerType)]),
		OwnerId:   tag.OwnerID.String(),
	}
}

func PbTagsToTags(tags []*pb.Tag) []*Tag {
	var ts []*Tag
	for _, tag := range tags {
		ts = append(ts, PbTagToTag(tag))
	}
	return ts
}

func TagsToPbTags(tags []*Tag) []*pb.Tag {
	var ts []*pb.Tag
	for _, tag := range tags {
		ts = append(ts, TagToPbTag(tag))
	}
	return ts
}

func PbLabelToLabel(label *pb.Label) *Label {
	id, err := uuid.Parse(label.GetId())
	if err != nil {
		id = uuid.New()
	}
	return &Label{
		ID:        id,
		Name:      label.GetName(),
		CreatedBy: label.GetCreatedBy(),
		OwnerType: OwnerType(pb.OwnerType_name[int32(label.GetOwnerType())]),
		OwnerID:   uuid.MustParse(label.GetOwnerId()),
	}
}

func LabelToPbLabel(label *Label) *pb.Label {
	return &pb.Label{
		Id:        label.ID.String(),
		Name:      label.Name,
		CreatedBy: label.CreatedBy,
		CreatedAt: label.CreatedAt.String(),
		UpdatedAt: label.UpdatedAt.String(),
		OwnerType: pb.OwnerType(pb.OwnerType_value[string(label.OwnerType)]),
		OwnerId:   label.OwnerID.String(),
	}
}

func PbLabelsToLabels(labels []*pb.Label) []*Label {
	var ls []*Label
	for _, label := range labels {
		ls = append(ls, PbLabelToLabel(label))
	}
	return ls
}

func LabelsToPbLabels(labels []*Label) []*pb.Label {
	var ls []*pb.Label
	for _, label := range labels {
		ls = append(ls, LabelToPbLabel(label))
	}
	return ls
}

func ContactCustomFieldValuesMaker(contactId uuid.UUID, customFieldDefId string, value string) *CustomFieldValue {
	newCustomFieldValue := &CustomFieldValue{
		ID:                      uuid.New(),
		ContactID:               contactId,
		CustomFieldDefinitionID: customFieldDefId,
		Value:                   value,
	}
	return newCustomFieldValue
}

func CustomFieldDefinitionToPbCustomFieldDefinition(customFieldDefinition *CustomFieldDefinition) *pb.CustomFieldDefinition {
	return &pb.CustomFieldDefinition{
		Id:        customFieldDefinition.ID.String(),
		Name:      customFieldDefinition.Name,
		Type:      customFieldDefinition.Type,
		OwnerType: pb.OwnerType(pb.OwnerType_value[string(customFieldDefinition.OwnerType)]),
		OwnerId:   customFieldDefinition.OwnerID.String(),
		CreatedAt: customFieldDefinition.CreatedAt.String(),
		UpdatedAt: customFieldDefinition.UpdatedAt.String(),
	}
}

func CustomFieldDefinitionsToPbCustomFieldDefinitions(customFieldDefinitions []*CustomFieldDefinition) []*pb.CustomFieldDefinition {
	var cfs []*pb.CustomFieldDefinition
	for _, customFieldDefinition := range customFieldDefinitions {
		cfs = append(cfs, CustomFieldDefinitionToPbCustomFieldDefinition(customFieldDefinition))
	}
	return cfs
}

// Function to safely retrieve a value from the record array
func ReadRecordValue(record []string, columnIndex map[string]int, columnName string) string {
	idx, ok := columnIndex[columnName]
	if !ok {
		return ""
	}
	return record[idx]
}

func PbContactGroupToContactGroup(contactGroup *pb.Group) *ContactGroup {
	if contactGroup == nil {
		return nil
	}

	id, err := uuid.Parse(contactGroup.GetId())
	if err != nil {
		id = uuid.New()
	}
	ownerID, err := uuid.Parse(contactGroup.GetOwnerId())
	if err != nil {
		ownerID = uuid.Nil
	}

	return &ContactGroup{
		ID:           id,
		SerialNumber: utils.CreateSerialNumber("CNTGP", id.String()),
		Name:         contactGroup.GetName(),
		ContactIds:   contactGroup.GetContactIds(),
		OwnerID:      ownerID,
		OwnerType:    OwnerType(contactGroup.GetOwnerType()),
	}
}

func ContactGroupToPbContactGroup(contactGroup *ContactGroup) *pb.Group {
	return &pb.Group{
		Id:         contactGroup.ID.String(),
		Name:       contactGroup.Name,
		ContactIds: contactGroup.ContactIds,
		CreatedAt:  contactGroup.CreatedAt.String(),
		UpdatedAt:  contactGroup.UpdatedAt.String(),
		OwnerId:    contactGroup.OwnerID.String(),
		OwnerType:  pb.OwnerType(pb.OwnerType_value[string(contactGroup.OwnerType)]),
	}
}

func PbContactGroupsToContactGroups(contactGroups []*pb.Group) []*ContactGroup {
	var cgs []*ContactGroup
	for _, contactGroup := range contactGroups {
		cgs = append(cgs, PbContactGroupToContactGroup(contactGroup))
	}
	return cgs
}

func ContactGroupsToPbContactGroups(contactGroups []*ContactGroup) []*pb.Group {
	var cgs []*pb.Group
	for _, contactGroup := range contactGroups {
		cgs = append(cgs, ContactGroupToPbContactGroup(contactGroup))
	}
	return cgs
}

func ConvertModelName(name string) string {
	// convert from snake_case to CamelCase
	// split the name into words
	words := strings.Split(name, "_")
	// convert the first letter of each word to uppercase
	for i, word := range words {
		words[i] = strings.ToTitle(word)
	}
	// join the words back together
	return strings.Join(words, "")
}

func PbSocialProfileToSocialProfile(profile *pb.SocialProfile) *SocialProfile {
	id, err := uuid.Parse(profile.GetId())
	if err != nil {
		id = uuid.New()
	}
	contactID, err := uuid.Parse(profile.GetContactId())
	if err != nil {
		contactID = uuid.Nil
	}
	return &SocialProfile{
		ID:         id,
		ContactID:  contactID,
		PlatformID: profile.GetPlatformId(),
		PageID:     profile.GetPageId(),
		ExternalID: profile.GetExternalId(),
	}
}

func SocialProfileToPbSocialProfile(profile *SocialProfile) *pb.SocialProfile {
	return &pb.SocialProfile{
		Id:         profile.ID.String(),
		ContactId:  profile.ContactID.String(),
		PlatformId: profile.PlatformID,
		PageId:     profile.PageID,
		ExternalId: profile.ExternalID,
	}
}

func PbSocialProfilesToSocialProfiles(profiles []*pb.SocialProfile) []*SocialProfile {
	var ps []*SocialProfile
	for _, profile := range profiles {
		ps = append(ps, PbSocialProfileToSocialProfile(profile))
	}
	return ps
}

func SocialProfilesToPbSocialProfiles(profiles []*SocialProfile) []*pb.SocialProfile {
	var ps []*pb.SocialProfile
	for _, profile := range profiles {
		ps = append(ps, SocialProfileToPbSocialProfile(profile))
	}
	return ps
}
