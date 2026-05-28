package casbin

import "fmt"

// SeedOrgAdmin seeds org-admin policies for the creator of a new organisation.
// It grants "user:{userID}" the "role:org-admin" role in "org:{orgID}" and
// adds a wildcard policy so the org-admin can do anything within the org domain.
func SeedOrgAdmin(e *Enforcer, orgID, userID string) error {
	domain := fmt.Sprintf("org:%s", orgID)
	role := "role:org-admin"

	// Wildcard policy for org-admin in this org.
	if err := e.AddPolicy(role, domain, "*", "*"); err != nil {
		return fmt.Errorf("casbin seed org-admin policy: %w", err)
	}

	// Assign the creator to org-admin in the org domain.
	if err := e.GrantRoleInDomain(userID, role, domain); err != nil {
		return fmt.Errorf("casbin seed org-admin role grant: %w", err)
	}

	return nil
}

// SeedWorkspaceOwner seeds workspace-owner policies for the creator of a new workspace.
// It grants "user:{userID}" the "role:ws-owner" role in "ws:{wsID}" with a wildcard
// policy, and also cross-links the workspace domain into the org domain so that org-admin
// can also manage the workspace.
func SeedWorkspaceOwner(e *Enforcer, wsID, orgID, userID, template string) error {
	wsDomain := fmt.Sprintf("ws:%s", wsID)
	orgDomain := fmt.Sprintf("org:%s", orgID)
	ownerRole := "role:ws-owner"
	memberRole := "role:ws-member"

	// Wildcard policy for ws-owner in this workspace.
	if err := e.AddPolicy(ownerRole, wsDomain, "*", "*"); err != nil {
		return fmt.Errorf("casbin seed ws-owner policy: %w", err)
	}

	// Wildcard policy for ws-admin — same privilege as owner for Phase 1.
	if err := e.AddPolicy("role:ws-admin", wsDomain, "*", "*"); err != nil {
		return fmt.Errorf("casbin seed ws-admin policy: %w", err)
	}

	// Org-admin inherits ws-owner in this workspace domain.
	if _, err := e.e.AddRoleForUserInDomain("role:org-admin", ownerRole, wsDomain); err != nil {
		return fmt.Errorf("casbin seed org-admin ws inheritance: %w", err)
	}
	_ = orgDomain // org domain referenced for clarity; inheritance is via role chain

	// Seed template-specific member policies.
	memberPolicies := templateMemberPolicies(template, wsDomain, memberRole)
	for _, mp := range memberPolicies {
		if err := e.AddPolicy(mp[0], mp[1], mp[2], mp[3]); err != nil {
			return fmt.Errorf("casbin seed ws-member policy (%v): %w", mp, err)
		}
	}

	// Assign the creator the ws-owner role in the workspace domain.
	if err := e.GrantRoleInDomain(userID, ownerRole, wsDomain); err != nil {
		return fmt.Errorf("casbin seed ws-owner role grant: %w", err)
	}

	return nil
}

// templateMemberPolicies returns the default ws-member policies for a given
// domain template. Returns [sub, dom, obj, act] tuples.
func templateMemberPolicies(template, wsDomain, memberRole string) [][4]string {
	switch template {
	case "support":
		return [][4]string{
			{memberRole, wsDomain, "ticket", "create"},
			{memberRole, wsDomain, "ticket", "read"},
			{memberRole, wsDomain, "ticket", "update"},
			{memberRole, wsDomain, "conversation", "read"},
			{memberRole, wsDomain, "knowledge_base", "read"},
			{memberRole, wsDomain, "sla", "read"},
		}
	case "sales":
		return [][4]string{
			{memberRole, wsDomain, "lead", "create"},
			{memberRole, wsDomain, "lead", "read"},
			{memberRole, wsDomain, "lead", "update"},
			{memberRole, wsDomain, "contact", "create"},
			{memberRole, wsDomain, "contact", "read"},
			{memberRole, wsDomain, "pipeline", "read"},
		}
	case "marketing":
		return [][4]string{
			{memberRole, wsDomain, "contact", "read"},
			{memberRole, wsDomain, "automation", "read"},
		}
	case "operations":
		return [][4]string{
			{memberRole, wsDomain, "shift", "read"},
			{memberRole, wsDomain, "leave_request", "create"},
			{memberRole, wsDomain, "leave_request", "read"},
		}
	default:
		// Custom template — no default member policies.
		return nil
	}
}

// SeedWorkspaceMember grants the correct workspace Casbin role to a user when
// they are added to a workspace via AddMember. Called after DB persist.
func SeedWorkspaceMember(e *Enforcer, wsID, userID, role string) error {
	wsDomain := fmt.Sprintf("ws:%s", wsID)
	casbinRole := WorkspaceRoleToPolicy(role)
	if err := e.GrantRoleInDomain(userID, casbinRole, wsDomain); err != nil {
		return fmt.Errorf("casbin seed workspace member (%s → %s in %s): %w", userID, casbinRole, wsDomain, err)
	}
	return nil
}

// RevokeWorkspaceMember removes all Casbin workspace roles for a user when
// they are removed from a workspace. Called after DB delete.
func RevokeWorkspaceMember(e *Enforcer, wsID, userID string) error {
	wsDomain := fmt.Sprintf("ws:%s", wsID)
	if err := e.RevokeAllRolesInDomain(userID, wsDomain); err != nil {
		return fmt.Errorf("casbin revoke workspace member (%s from %s): %w", userID, wsDomain, err)
	}
	return nil
}

// WorkspaceRoleToPolicy maps a WorkspaceMember role string to the corresponding
// Casbin role name. Defaults to "role:ws-member" for unknown values.
func WorkspaceRoleToPolicy(role string) string {
	switch role {
	case "owner":
		return "role:ws-owner"
	case "admin":
		return "role:ws-admin"
	case "guest":
		return "role:ws-guest"
	default: // "member" + unknown
		return "role:ws-member"
	}
}
