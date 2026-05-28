package workspace

import (
	"gorm.io/gorm"
)

// WorkspaceStore handles all database operations for Workspace and WorkspaceMember.
type WorkspaceStore struct {
	db *gorm.DB
}

// NewWorkspaceStore returns a new WorkspaceStore backed by db.
func NewWorkspaceStore(db *gorm.DB) *WorkspaceStore {
	return &WorkspaceStore{db: db}
}

// Create persists a new Workspace.
func (s *WorkspaceStore) Create(w *Workspace) (*Workspace, error) {
	if err := s.db.Create(w).Error; err != nil {
		return nil, err
	}
	return w, nil
}

// FindByID retrieves a Workspace by its primary key.
func (s *WorkspaceStore) FindByID(id string) (*Workspace, error) {
	var w Workspace
	if err := s.db.First(&w, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// FindByOrgID retrieves all Workspaces belonging to an organisation.
func (s *WorkspaceStore) FindByOrgID(orgID string) ([]*Workspace, error) {
	var workspaces []*Workspace
	if err := s.db.Where("org_id = ? AND status != ?", orgID, WorkspaceStatusArchived).
		Order("created_at ASC").
		Find(&workspaces).Error; err != nil {
		return nil, err
	}
	return workspaces, nil
}

// Update saves changes to an existing Workspace.
func (s *WorkspaceStore) Update(w *Workspace) (*Workspace, error) {
	if err := s.db.Omit("ID", "OrgID", "CreatedAt").Save(w).Error; err != nil {
		return nil, err
	}
	return w, nil
}

// Archive sets a workspace status to archived.
func (s *WorkspaceStore) Archive(id string) error {
	return s.db.Model(&Workspace{}).
		Where("id = ?", id).
		Update("status", WorkspaceStatusArchived).Error
}

// AddMember creates a WorkspaceMember record.
func (s *WorkspaceStore) AddMember(m *WorkspaceMember) (*WorkspaceMember, error) {
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// RemoveMember deletes a WorkspaceMember by workspace+user pair.
func (s *WorkspaceStore) RemoveMember(workspaceID, userID string) error {
	return s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&WorkspaceMember{}).Error
}

// UpdateMemberRole changes the role of a workspace member.
func (s *WorkspaceStore) UpdateMemberRole(workspaceID, userID, role string) (*WorkspaceMember, error) {
	var m WorkspaceMember
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&m, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error; err != nil {
			return err
		}
		m.Role = MemberRole(role)
		return tx.Save(&m).Error
	})
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMembers returns all WorkspaceMembers for a workspace.
func (s *WorkspaceStore) ListMembers(workspaceID string) ([]*WorkspaceMember, error) {
	var members []*WorkspaceMember
	if err := s.db.Where("workspace_id = ?", workspaceID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// IsMember checks whether a user belongs to a workspace and returns their role.
func (s *WorkspaceStore) IsMember(workspaceID, userID string) (bool, MemberRole, error) {
	var m WorkspaceMember
	err := s.db.First(&m, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error
	if err == gorm.ErrRecordNotFound {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, m.Role, nil
}
