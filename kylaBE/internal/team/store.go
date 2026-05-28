package team

import (
	"kyla-be/internal/authctx"
	"kyla-be/pkg/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// TeamStore is the database access layer for teams.
// The DB field is exported so that external packages can access it directly.
type TeamStore struct {
	DB *gorm.DB
}

// NewTeamStore creates a new TeamStore.
func NewTeamStore(db *gorm.DB) *TeamStore {
	return &TeamStore{
		DB: db,
	}
}

func (d *TeamStore) CreateTeam(tm *Team) (*Team, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tm).Error; err != nil {
			return status.Error(500, "Internal Server Error, Failed to create team")
		}
		if err := tx.Model(tm).Association("Roles").Replace(tm.Roles); err != nil {
			return status.Error(500, "Internal Server Error, Failed to associate roles with team")
		}
		if err := tx.Model(tm).Association("Users").Replace(tm.Users); err != nil {
			return status.Error(500, "Internal Server Error, Failed to associate users with team")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tm, nil
}

func (d *TeamStore) ReadTeam(id *uuid.UUID) (*Team, error) {
	var tm Team
	if err := d.DB.First(&tm, id).Error; err != nil {
		return nil, err
	}
	return &tm, nil
}

func (d *TeamStore) ReadTeamsByUserID(userID string) ([]*Team, error) {
	var teams []*Team
	if err := d.DB.Joins("JOIN team_users ON team_users.team_id = teams.id").Where("team_users.user_id = ?", userID).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) ReadTeamsByOrganisationID(organisationID string) ([]*Team, error) {
	var teams []*Team
	if err := d.DB.Where("owner_id = ? AND owner_type = ?", organisationID, authctx.ORGANISATIONS).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) ReadTeamsByBranchID(branchID string) ([]*Team, error) {
	var teams []*Team
	if err := d.DB.Where("owner_id = ? AND owner_type = ?", branchID, authctx.BRANCHES).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) ReadTeamsByDepartmentID(departmentID string) ([]*Team, error) {
	var teams []*Team
	if err := d.DB.Where("owner_id = ? AND owner_type = ?", departmentID, authctx.DEPARTMENTS).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) UpdateTeam(tm *Team) (*Team, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(
			"created_at",
			"id",
			"serial_number",
			"owner_id",
			"owner_type",
		).Save(tm).Error; err != nil {
			return status.Error(500, "Internal Server Error, Failed to update team")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tm, nil
}

func (d *TeamStore) DeleteTeam(id uuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).Delete(&Team{}).Error
	})
}

// AddUserToTeam adds a user to a team via the user_teams join table.
func (d *TeamStore) AddUserToTeam(teamID, userID uuid.UUID) error {
	var tm Team
	if err := d.DB.First(&tm, "id = ?", teamID).Error; err != nil {
		return status.Error(404, "Not Found, Team not found")
	}

	userRef := UserRef{ID: userID}
	if err := d.DB.Model(&tm).Association("Users").Append(&userRef); err != nil {
		return status.Error(500, "Internal Server Error, Failed to add user to team")
	}
	return nil
}

// RemoveUserFromTeam removes a user from a team via the user_teams join table.
func (d *TeamStore) RemoveUserFromTeam(teamID, userID uuid.UUID) error {
	var tm Team
	if err := d.DB.First(&tm, "id = ?", teamID).Error; err != nil {
		return status.Error(404, "Not Found, Team not found")
	}

	userRef := UserRef{ID: userID}
	if err := d.DB.Model(&tm).Association("Users").Delete(&userRef); err != nil {
		return status.Error(500, "Internal Server Error, Failed to remove user from team")
	}
	return nil
}

// ReadTeamUsers returns the full User records for all members of a team.
// service.User is used here (not UserRef) because callers need the complete
// user data for the gRPC response; pkg/service does not import internal/team
// so there is no import cycle.
func (d *TeamStore) ReadTeamUsers(teamID uuid.UUID, roleID uuid.UUID) ([]service.User, error) {
	var users []service.User
	query := d.DB.Table("users").Joins("JOIN team_users ON team_users.user_id = users.id")

	if roleID != uuid.Nil {
		query = query.Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Where("team_users.team_id = ? AND user_roles.role_id = ?", teamID, roleID)
	} else {
		query = query.Where("team_users.team_id = ?", teamID)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, status.Error(500, "Internal Server Error, Failed to fetch team users")
	}
	return users, nil
}

func (d *TeamStore) ReadTeams(scope *authctx.OpScope) ([]*Team, error) {
	var teams []*Team
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("owner_id = ? AND owner_type = ?", scope.ID, scope.Owner).Find(&teams).Error
	})

	return teams, err
}
