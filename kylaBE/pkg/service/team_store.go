package service

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type TeamStore struct {
	DB *gorm.DB
}

// NewBranchStore creates a new BranchStore
func NewTeamStore(db *gorm.DB) *TeamStore {
	return &TeamStore{
		DB: db,
	}
}

func (d *TeamStore) CreateTeam(team *Team) (*Team, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return status.Error(500, "Internal Server Error, Failed to create team")
		}
		if err := tx.Model(team).Association("Roles").Replace(team.Roles); err != nil {
			return status.Error(500, "Internal Server Error, Failed to associate roles with team")
		}
		// Assuming team.Users is a slice of users to be associated with the team
		if err := tx.Model(team).Association("Users").Replace(team.Users); err != nil {
			return status.Error(500, "Internal Server Error, Failed to associate users with team")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (d *TeamStore) ReadTeam(id *uuid.UUID) (*Team, error) {
	var team Team
	if err := d.DB.First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
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
	if err := d.DB.Where("owner_id = ? AND owner_type = ?", organisationID, ORGANISATIONS).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) ReadTeamsByBranchID(branchID string) ([]*Team, error) {
	var teams []*Team
	if err := d.DB.Where("owner_id = ? AND owner_type = ?", branchID, BRANCHES).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) ReadTeamsByDepartmentID(departmentID string) ([]*Team, error) {
	var teams []*Team
	if err := d.DB.Where("owner_id = ? AND owner_type = ?", departmentID, DEPARTMENTS).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *TeamStore) UpdateTeam(team *Team) (*Team, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(
			"created_at",
			"id",
			"serial_number",
			"owner_id",
			"owner_type",
		).Save(team).Error; err != nil {
			return status.Error(500, "Internal Server Error, Failed to update team")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (d *TeamStore) DeleteTeam(id uuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).Delete(&Team{}).Error
	})
}

// add user to team
func (d *TeamStore) AddUserToTeam(teamID, userID uuid.UUID) error {
	var team Team
	if err := d.DB.First(&team, "id = ?", teamID).Error; err != nil {
		return status.Error(404, "Not Found, Team not found")
	}

	var user User
	if err := d.DB.First(&user, "id = ?", userID).Error; err != nil {
		return status.Error(404, "Not Found, User not found")
	}

	if err := d.DB.Model(&team).Association("Users").Append(&user); err != nil {
		return status.Error(500, "Internal Server Error, Failed to add user to team")
	}
	return nil
}

// remove user from team
func (d *TeamStore) RemoveUserFromTeam(teamID, userID uuid.UUID) error {
	var team Team
	if err := d.DB.First(&team, "id = ?", teamID).Error; err != nil {
		return status.Error(404, "Not Found, Team not found")
	}

	var user User
	if err := d.DB.First(&user, "id = ?", userID).Error; err != nil {
		return status.Error(404, "Not Found, User not found")
	}

	if err := d.DB.Model(&team).Association("Users").Delete(&user); err != nil {
		return status.Error(500, "Internal Server Error, Failed to remove user from team")
	}
	return nil
}

func (d *TeamStore) ReadTeamUsers(teamID uuid.UUID, roleID uuid.UUID) ([]User, error) {
	var users []User
	query := d.DB.Joins("JOIN team_users ON team_users.user_id = users.id")

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

func (d *TeamStore) ReadTeams(scope *OpScope) ([]*Team, error) {
	var teams []*Team
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("owner_id = ? AND owner_type = ?", scope.ID, scope.Owner).Find(&teams).Error
	})

	return teams, err
}
