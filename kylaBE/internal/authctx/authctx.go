package authctx

import "github.com/google/uuid"

// Context key type to avoid collisions with other packages.
type key int

const (
	UserID         key = iota
	OrganisationID key = iota
	Authorization  key = iota
	RequestAuth    key = iota
	BranchID       key = iota
	Roles          key = iota
	Scope          key = iota
	WorkspaceID    key = iota // Active workspace scope (extracted from x-workspace-id gRPC header)
)

// OwnerType identifies the ownership entity type for roles and resources.
type OwnerType string

const (
	USERS         OwnerType = "USERS"
	TEAMS         OwnerType = "TEAMS"
	BRANCHES      OwnerType = "BRANCHES"
	DEPARTMENTS   OwnerType = "DEPARTMENTS"
	ORGANISATIONS OwnerType = "ORGANISATIONS"
)

// OpScope is a typed scope check descriptor used in context validation.
type OpScope struct {
	Owner OwnerType
	ID    string
}

// IdNameMapping is a lightweight ID→Name pair used in scope resolution.
type IdNameMapping struct {
	ID   string
	Name string
}

// Scopes holds the ownership boundaries resolved for a request.
type Scopes struct {
	User         uuid.UUID
	Teams        []string
	Departments  []string
	Branches     []string
	Branch       uuid.UUID // Current branch
	Organisation uuid.UUID
}

// RequestMetadata is the resolved auth context attached to every gRPC request.
type RequestMetadata struct {
	Authorization  string
	OrganisationID uuid.UUID
	UserID         uuid.UUID
	BranchID       uuid.UUID
	WorkspaceID    uuid.UUID // Active workspace scope (zero-value when not set)
	RequestAuth    string
	Roles          []uuid.UUID
	Scopes         *Scopes
	IdNameMappings []*IdNameMapping
}
