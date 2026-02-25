package service

import (
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DatabaseUserStore is a database implementation of the UserStore
type UserStore struct {
	db *gorm.DB
}

// NewUserStore creates a new database user store
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

// Save persists a user to the database
func (store *UserStore) Save(user *User, new bool) (*User, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if new {
			if err := tx.Set("gorm:save_Associations", false).Create(user).Error; err != nil {
				return fmt.Errorf("failed to save user: %v", err)
			}
		} else {
			if err := tx.Model(user).Association("Roles").Replace(user.Roles); err != nil {
				return fmt.Errorf("failed to save user roles: %v", err)
			}
			if err := tx.Save(user).Preload(clause.Associations).Find(user, "id = ?", user.ID).Error; err != nil {
				return fmt.Errorf("failed to find user: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByUsername Find by username
func (store *UserStore) FindByUsername(username string) (*User, error) {
	var user User
	result := store.db.Preload(clause.Associations).First(&user, "username = ?", username)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}
	return &user, nil
}

func (store *UserStore) FindByEmailOrUsername(email, username string) (*User, error) {
	var user User
	result := store.db.Preload(clause.Associations).First(&user, "email = ? OR username = ?", email, username)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}
	return &user, nil
}

func (store *UserStore) CheckUserExist(email, username string, phone string) bool {
	var user User
	// Match for email or username in the database and the other one is not empty
	emailResult := store.db.First(&user, "email = ?", email)
	usernameResult := store.db.First(&user, "username = ?", username)
	phoneResult := store.db.First(&user, "phone = ?", phone)

	return emailResult.Error == nil || usernameResult.Error == nil || phoneResult.Error == nil
}

// FindByEmail retrieves a user from the database based on email
func (store *UserStore) FindByEmail(email string) (*User, error) {
	var user User
	result := store.db.Preload(clause.Associations).First(&user, "email = ?", email)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by email: %v", result.Error)
	}
	return &user, nil
}

// FindAllByIds retrieves users from the database based on user IDs
func (store *UserStore) FindAllByIds(ids []string) ([]*User, error) {
	var users []*User
	result := store.db.Where("id IN ?", ids).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users by IDs: %v", result.Error)
	}
	return users, nil
}

// FindByPhone retrieves a user from the database based on phone
func (store *UserStore) FindByPhone(phone string) (*User, error) {
	var user User
	result := store.db.First(&user, "phone = ?", phone)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by phone: %v", result.Error)
	}
	return &user, nil
}

// FindByDepartments retrieves users from the database based on department IDs
func (store *UserStore) FindByDepartments(departmentIDs []string) ([]*User, error) {
	var users []*User
	result := store.db.Where("departments @> ?", pq.StringArray(departmentIDs)).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users by departments: %v", result.Error)
	}
	return users, nil
}

// Update updates the user in the database
func (store *UserStore) Update(user *User) (*User, error) {
	result := store.db.Omit(
		"CreatedAt",
		"CreatedBy",
		"Status",
		"Roles",
		"CurrentOrganisationID",
		"CurrentBranchID",
		"ID",
		"AgentStatusID",
		"BranchID",
		"IsDefault",
		"SerialNumber",
		"Username",
		"Email",
		"OwnerType",
		"OwnerID",
	).Save(user).Find(user, "id = ?", user.ID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update user: %v", result.Error)
	}
	return user, nil
}

// Delete deletes the user from the database
func (store *UserStore) Delete(userID string) error {
	result := store.db.Where("id = ?", userID).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %v", result.Error)
	}
	return nil
}

func (store *UserStore) FindByScope(scope *OpScope) ([]*User, error) {
	var users []*User
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where(("owner_id = ? AND owner_type = ?"), scope.ID, scope.Owner).Find(&users).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find users by scope: %v", err)
	}
	return users, nil
}

// FindByID retrieves a user from the database based on ID
func (store *UserStore) FindByID(id *uuid.UUID) (*User, error) {
	var user User
	result := store.db.Preload(clause.Associations).First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", result.Error)
	}
	return &user, nil
}

// FindAll retrieves all users from the database
func (store *UserStore) FindAll() ([]*User, error) {
	var users []*User
	result := store.db.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all users: %v", result.Error)
	}
	return users, nil
}

// FindUserByResetToken retrieves a user from the database based on reset token
func (store *UserStore) FindUserByResetToken(resetToken string) (*User, error) {
	var user User
	result := store.db.Where("reset_token = ?", resetToken).First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by reset token: %v", result.Error)
	}
	return &user, nil
}

// FindUserByForgotPasswordRequest retrieves a user from the database based on ForgotPasswordRequest
func (store *UserStore) FindUserByForgotPasswordRequest(req *pb.ForgotPasswordRequest) (*User, error) {
	var user User

	switch identity := req.GetIdentity().(type) {
	case *pb.ForgotPasswordRequest_UserId:
		// Find user by user ID
		result := store.db.First(&user, "id = ?", identity.UserId)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find user by user ID: %v", result.Error)
		}

	case *pb.ForgotPasswordRequest_Email:
		// Find user by email
		result := store.db.First(&user, "email = ?", identity.Email)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find user by email: %v", result.Error)
		}

	case *pb.ForgotPasswordRequest_Username:
		// Find user by username
		result := store.db.First(&user, "username = ?", identity.Username)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find user by username: %v", result.Error)
		}

	default:
		return nil, fmt.Errorf("unsupported identity type in request")
	}

	return &user, nil
}

// ReadUserPasswordByID retrieves the hashed password of a user by user ID
func (store *UserStore) ReadUserPasswordByID(userID string) (string, error) {
	var user User
	result := store.db.First(&user, "id = ?", userID)
	if result.Error != nil {
		return "", fmt.Errorf("failed to get user password: %v", result.Error)
	}
	return user.HashedPassword, nil
}

func (store *UserStore) UpdateUserPassword(userID, newPassword string) error {
	var user User
	result := store.db.Model(&user).Where("id = ?", userID).Update("hashed_password", newPassword)
	if result.Error != nil {
		return fmt.Errorf("failed to update user password: %v", result.Error)
	}
	return nil
}

// ChangePassword updates the user's password in the database
func (store *UserStore) ChangePassword(userID *uuid.UUID, newPassword string) (*User, error) {
	// Read the user from the database
	user, err := store.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}

	user.HashedPassword = utils.HASH_PASSWORD(newPassword)
	// Save the updated user by updating only the password field
	if _, err := store.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}
	return user, nil
}

// Retrieve all users in the db
func (store *UserStore) RetrieveAllUsers() ([]*User, error) {
	var users []*User
	result := store.db.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all users: %v", result.Error)
	}
	return users, nil
}

// FindUsersByRoleIds retrieves all users from the database based on role IDs
func (store *UserStore) FindUsersByRoleIds(roleIds []string) ([]*User, error) {
	var users []*User
	err := store.db.Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id IN ?", roleIds).
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by role IDs: %v", err)
	}
	return users, nil
}

// AddRoleToUser adds a role to a user in the database
func (store *UserStore) AddRoleToUser(userID, roleID string) (*User, error) {
	// Read the user from the database
	user := &User{}
	err := store.db.Preload("Roles").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}

	// Read the role from the database
	role := &Role{}
	err = store.db.First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find role by ID: %v", err)
	}

	// Add the role to the user
	err = store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Association("Roles").Append(role)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add role to user: %v", err)
	}
	return user, nil
}

func (store *UserStore) RemoveRoleFromUser(userID, roleID string) (*User, error) {
	// Read the user from the database
	user := &User{}
	err := store.db.Preload("Roles").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}
	roles := make([]Role, 0)
	for _, r := range user.Roles {
		if r.ID.String() != roleID {
			roles = append(roles, r)
		}
	}
	err = store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Association("Roles").Replace(roles)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove role from user: %v", err)
	}
	return user, nil
}

func (store *UserStore) AddUsersToRole(roleID string, userIds []string) error {
	users, err := store.FindAllByIds(userIds)
	if err != nil {
		return fmt.Errorf("failed to find users by IDs: %v", err)
	}
	for _, user := range users {
		if _, err := store.AddRoleToUser(user.ID.String(), roleID); err != nil {
			return fmt.Errorf("failed to add role to user: %v", err)
		}
	}
	return nil
}

func (store *UserStore) DeactivateUser(userID *uuid.UUID) error {
	user, err := store.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to find user by ID: %v", err)
	}
	user.Status = k.USER_STATUSES()["INACTIVE"]
	if _, err := store.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

func (store *UserStore) ActivateUser(userID *uuid.UUID) error {
	user, err := store.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to find user by ID: %v", err)
	}
	user.Status = k.USER_STATUSES()["ACTIVE"]
	if _, err := store.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

func (store *UserStore) Search(query string, page int32, perPage int32) ([]*User, int32, error) {
	var users []*User
	var count int64
	offset := (page - 1) * perPage

	result := store.db.Where("(first_name ILIKE ? OR last_name ILIKE ? OR username ILIKE ? OR email ILIKE ? OR phone ILIKE ?)", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").Find(&users).Count(&count).Offset(int(offset)).Limit(int(perPage))
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to search users: %v", result.Error)
	}
	return users, int32(count), nil
}

// FindByFirebaseUID retrieves a user from the database based on Firebase UID
func (store *UserStore) FindByFirebaseUID(firebaseUID string) (*User, error) {
	var user User
	result := store.db.Preload(clause.Associations).First(&user, "firebase_uid = ?", firebaseUID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by Firebase UID: %v", result.Error)
	}
	return &user, nil
}

// Create creates a new user in the database
func (store *UserStore) Create(user *User) (*User, error) {
	result := store.db.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}
	return user, nil
}

func (store *UserStore) FindByIdsFromService(ids []string) ([]*User, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}

	var users []*User
	result := store.db.Preload(clause.Associations).Where("id IN (?)", ids).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users by IDs: %v", result.Error)
	}

	return users, nil
}
