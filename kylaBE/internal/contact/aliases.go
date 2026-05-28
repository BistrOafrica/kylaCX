// Package contact provides access to the contact domain types.
// Phase 0 bridge: re-exports from pkg/service with forwarding constructors.
// Phase 1 will replace these with native implementations.
package contact

import (
	"kyla-be/internal/auth"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/service"

	"gorm.io/gorm"
)

// Type aliases — exact re-exports of the service layer types.
type Contact = service.Contact
type SocialProfile = service.SocialProfile
type ContactGroup = service.ContactGroup
type CustomFieldDefinition = service.CustomFieldDefinition
type CustomFieldValue = service.CustomFieldValue
type CustomField = service.CustomField
type ContactMergeHistory = service.ContactMergeHistory
type ContactStore = service.ContactStore
type ContactGroupStore = service.ContactGroupStore
type BranchStore = service.BranchStore
type SharingStore = service.SharingStore
type Contacts = service.Contacts
type Groups = service.Groups

// NewContactStore creates a new ContactStore.
func NewContactStore(db *gorm.DB) *ContactStore {
	return service.NewContactStore(db)
}

// NewContactGroupStore creates a new ContactGroupStore.
func NewContactGroupStore(db *gorm.DB) *ContactGroupStore {
	return service.NewContactGroupStore(db)
}

// NewBranchStore creates a new BranchStore (service-layer alias).
func NewBranchStore(db *gorm.DB) *BranchStore {
	return service.NewBranchStore(db)
}

// NewSharingStore creates a new SharingStore (service-layer alias).
func NewSharingStore(db *gorm.DB) *SharingStore {
	return service.NewSharingStore(db)
}

// NewContactServer creates a new Contacts gRPC server.
func NewContactServer(
	contactStore *ContactStore,
	authStore *auth.AuthStore,
	branchStore *BranchStore,
	groupStore *ContactGroupStore,
	leadClient pb.LeadServiceClient,
	sharingStore *SharingStore,
) *Contacts {
	return service.NewContactServer(contactStore, authStore.Inner(), branchStore, groupStore, leadClient, sharingStore)
}

// NewContactGroupServer creates a new Groups gRPC server.
func NewContactGroupServer(
	groupStore *ContactGroupStore,
	authStore *auth.AuthStore,
	contactStore *ContactStore,
) *Groups {
	return service.NewContactGroupServer(groupStore, authStore.Inner(), contactStore)
}
