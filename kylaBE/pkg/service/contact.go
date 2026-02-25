package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model
	ID                   uuid.UUID          `gorm:"primarykey;type:uuid;not null; index"`
	FirstName            string             `gorm:"not null; index"`
	LastName             string             `gorm:"index"`
	OtherName            string             `gorm:"index"`
	Email                string             `gorm:"column:email; index"`
	Phone                string             `gorm:"column:phone; index"`
	OtherPhone           string             `gorm:"column:other_phone; index"`
	SerialNumber         string             `gorm:"index"`
	GroupIds             []string           `gorm:"type:uuid[];allowNull:true;default:null"`
	Title                string             `gorm:"column:title"`
	Prefix               string             `gorm:"column:prefix"`
	Suffix               string             `gorm:"column:suffix"`
	JobDepartment        string             `gorm:"column:job_department"`
	JobTitle             string             `gorm:"column:job_title"`
	Company              string             `gorm:"column:company; index"`
	Nickname             string             `gorm:"column:nickname; index"`
	Notes                string             `gorm:"column:notes"`
	Birthday             string             `gorm:"column:birthday"`
	Country              string             `gorm:"column:country; index"`
	State                string             `gorm:"column:state; index"`
	City                 string             `gorm:"column:city; index"`
	Street               string             `gorm:"column:street; index"`
	PostalCode           string             `gorm:"column:postal_code; index"`
	URL                  string             `gorm:"column:url; index"`
	SocialProfiles       []SocialProfile    `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE"`
	Tags                 []Tag              `gorm:"many2many:contact_tags"`
	Labels               []Label            `gorm:"many2many:contact_labels"`
	CreatedBy            string             `gorm:"column:created_by; default:null"`
	CustomFieldValues    []CustomFieldValue `gorm:"foreignKey:ContactID;references:ID"`
	PipelineIDs          pq.StringArray     `gorm:"type:text[]"`
	StageIDs             pq.StringArray     `gorm:"type:text[]"`
	OwnerID              uuid.UUID          `gorm:"not null; type:uuid; index"`
	OwnerType            OwnerType          `gorm:"not null; default:USERS; index"`
	ParentId             uuid.UUID          `gorm:"allowNull:true;default:null;type:uuid"`
	NormalizedPhone      string             `gorm:"column:normalized_phone; index"`
	NormalizedOtherPhone string             `gorm:"column:normalized_other_phone; index"`
	IsVip                bool               `gorm:"column:is_vip; default:false; index"`
}

type SocialProfile struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	ContactID  uuid.UUID `gorm:"not null;index"`
	PlatformID string    `gorm:"not null;index;index"`
	PageID     string    `gorm:"index;index"`
	ExternalID string    `gorm:"index;not null;index"`
}

type Tag struct {
	gorm.Model
	ID        uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	ColorCode string
	Name      string
	Contacts  []Contact `gorm:"many2many:contact_tags"`
	CreatedBy string
	OwnerID   uuid.UUID `gorm:"not null; type:uuid; index"`
	OwnerType OwnerType `gorm:"not null; default:USERS; index"`
}

type Label struct {
	gorm.Model
	ID        uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	Name      string
	Contacts  []Contact `gorm:"many2many:contact_labels"`
	CreatedBy string
	OwnerID   uuid.UUID `gorm:"not null; type:uuid; index"`
	OwnerType OwnerType `gorm:"not null; default:USERS; index"`
}

type ContactGroup struct {
	gorm.Model
	ID           uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	SerialNumber string
	Name         string
	ContactIds   pq.StringArray `gorm:"type:text[];default:null"`
	CreatedBy    string
	UpdatedBy    string
	OwnerID      uuid.UUID `gorm:"not null; type:uuid; index"`
	OwnerType    OwnerType `gorm:"not null; default:USERS; index"`
}

type CustomFieldDefinition struct {
	gorm.Model
	ID                uuid.UUID          `gorm:"primarykey;type:uuid;not null"`
	Name              string             `gorm:"not null"`
	Type              string             `gorm:"not null"`
	CustomFieldValues []CustomFieldValue `gorm:"foreignKey:CustomFieldDefinitionID;references:ID"`
	OwnerID           uuid.UUID          `gorm:"not null; type:uuid; index"`
	OwnerType         OwnerType          `gorm:"not null; default:USERS; index"`
}

type CustomFieldValue struct {
	gorm.Model
	ID                      uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	ContactID               uuid.UUID `gorm:"type:uuid;not null; uniqueIndex:idx_contact_custom_field"`
	CustomFieldDefinitionID string    `gorm:"not null; uniqueIndex:idx_contact_custom_field"`
	Value                   string    `gorm:"not null"`
}

type CustomField struct {
	Name         string
	Value        string
	Type         string
	DefinitionID string
	ValueID      string
	OwnerType    OwnerType
	OwnerID      string
}

type ContactMergeHistory struct {
	gorm.Model
	ID              uuid.UUID `gorm:"type:uuid;primary_key"`
	MasterContactID uuid.UUID `gorm:"type:uuid;not null"`
	MergedContactID uuid.UUID `gorm:"type:uuid;not null"`
	MergedAt        time.Time
	MergedBy        string `gorm:"default:null"`
}
