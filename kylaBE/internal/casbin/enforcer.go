package casbin

import (
	_ "embed"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

//go:embed model.conf
var modelConf string

// Enforcer wraps the Casbin RBAC-with-domains enforcer.
type Enforcer struct {
	e *casbin.Enforcer
}

// NewEnforcer constructs a Casbin enforcer backed by the provided GORM DB.
// The gorm-adapter automatically creates the "casbin_rules" table if it does not exist.
func NewEnforcer(db *gorm.DB) (*Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, err
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}

	return &Enforcer{e: e}, nil
}

// Enforce checks whether "user:{userID}" can perform act on obj within domain dom.
func (e *Enforcer) Enforce(userID, domain, obj, act string) (bool, error) {
	return e.e.Enforce("user:"+userID, domain, obj, act)
}

// GrantRoleInDomain assigns role to userID within domain.
// role should already be prefixed (e.g. "role:org-admin").
func (e *Enforcer) GrantRoleInDomain(userID, role, domain string) error {
	_, err := e.e.AddRoleForUserInDomain("user:"+userID, role, domain)
	return err
}

// RevokeRoleInDomain removes role from userID within domain.
func (e *Enforcer) RevokeRoleInDomain(userID, role, domain string) error {
	_, err := e.e.DeleteRoleForUserInDomain("user:"+userID, role, domain)
	return err
}

// AddPolicy adds a direct (sub, dom, obj, act) policy rule.
func (e *Enforcer) AddPolicy(sub, dom, obj, act string) error {
	_, err := e.e.AddPolicy(sub, dom, obj, act)
	return err
}

// RemovePolicy removes a (sub, dom, obj, act) policy rule.
func (e *Enforcer) RemovePolicy(sub, dom, obj, act string) error {
	_, err := e.e.RemovePolicy(sub, dom, obj, act)
	return err
}

// GetRolesInDomain returns all roles held by userID in domain.
func (e *Enforcer) GetRolesInDomain(userID, domain string) []string {
	return e.e.GetRolesForUserInDomain("user:"+userID, domain)
}

// RevokeAllRolesInDomain removes every role held by userID within domain.
// Use this when a member is removed from a workspace entirely.
func (e *Enforcer) RevokeAllRolesInDomain(userID, domain string) error {
	_, err := e.e.DeleteRolesForUserInDomain("user:"+userID, domain)
	return err
}

// LoadPolicy reloads all policies from the database.
// Call this after bulk external changes to the casbin_rules table.
func (e *Enforcer) LoadPolicy() error {
	return e.e.LoadPolicy()
}
