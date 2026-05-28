package user

import (
	"fmt"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserStore handles persistence for User records.
type UserStore struct {
	DB *gorm.DB
}

// NewUserStore creates a new UserStore.
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{
		DB: db,
	}
}

// Save persists a user. When new is true it creates; otherwise it replaces roles and saves.
func (store *UserStore) Save(user *User, new bool) (*User, error) {
	err := store.DB.Transaction(func(tx *gorm.DB) error {
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

// FindByUsername retrieves a user by username, preloading associations.
func (store *UserStore) FindByUsername(username string) (*User, error) {
	var user User
	result := store.DB.Preload(clause.Associations).First(&user, "username = ?", username)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}
	return &user, nil
}

// FindByEmailOrUsername retrieves a user matching either email or username.
func (store *UserStore) FindByEmailOrUsername(email, username string) (*User, error) {
	var user User
	result := store.DB.Preload(clause.Associations).First(&user, "email = ? OR username = ?", email, username)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %v", result.Error)
	}
	return &user, nil
}

// CheckUserExist returns true if a user with the given email, username, or phone already exists.
func (store *UserStore) CheckUserExist(email, username string, phone string) bool {
	var user User
	emailResult := store.DB.First(&user, "email = ?", email)
	usernameResult := store.DB.First(&user, "username = ?", username)
	phoneResult := store.DB.First(&user, "phone = ?", phone)
	return emailResult.Error == nil || usernameResult.Error == nil || phoneResult.Error == nil
}

// FindByEmail retrieves a user by email.
func (store *UserStore) FindByEmail(email string) (*User, error) {
	var user User
	result := store.DB.Preload(clause.Associations).First(&user, "email = ?", email)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by email: %v", result.Error)
	}
	return &user, nil
}

// FindAllByIds retrieves users by a slice of ID strings.
func (store *UserStore) FindAllByIds(ids []string) ([]*User, error) {
	var users []*User
	result := store.DB.Where("id IN ?", ids).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users by IDs: %v", result.Error)
	}
	return users, nil
}

// FindByPhone retrieves a user by phone number.
func (store *UserStore) FindByPhone(phone string) (*User, error) {
	var user User
	result := store.DB.First(&user, "phone = ?", phone)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by phone: %v", result.Error)
	}
	return &user, nil
}

// FindByDepartments retrieves users belonging to the specified department IDs.
func (store *UserStore) FindByDepartments(departmentIDs []string) ([]*User, error) {
	var users []*User
	result := store.DB.Where("departments @> ?", pq.StringArray(departmentIDs)).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users by departments: %v", result.Error)
	}
	return users, nil
}

// Update saves user changes, omitting fields that must not be overwritten.
func (store *UserStore) Update(user *User) (*User, error) {
	result := store.DB.Omit(
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

// Delete soft-deletes a user by ID string.
func (store *UserStore) Delete(userID string) error {
	result := store.DB.Where("id = ?", userID).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %v", result.Error)
	}
	return nil
}

// FindByScope retrieves users matching a given ownership scope.
func (store *UserStore) FindByScope(scope *authctx.OpScope) ([]*User, error) {
	var users []*User
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("owner_id = ? AND owner_type = ?", scope.ID, scope.Owner).Find(&users).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find users by scope: %v", err)
	}
	return users, nil
}

// FindByID retrieves a user by UUID, preloading associations.
func (store *UserStore) FindByID(id *uuid.UUID) (*User, error) {
	var user User
	result := store.DB.Preload(clause.Associations).First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", result.Error)
	}
	return &user, nil
}

// FindAll retrieves all users.
func (store *UserStore) FindAll() ([]*User, error) {
	var users []*User
	result := store.DB.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all users: %v", result.Error)
	}
	return users, nil
}

// FindUserByResetToken retrieves a user by password reset token.
func (store *UserStore) FindUserByResetToken(resetToken string) (*User, error) {
	var user User
	result := store.DB.Where("reset_token = ?", resetToken).First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by reset token: %v", result.Error)
	}
	return &user, nil
}

// FindUserByForgotPasswordRequest retrieves a user matching a ForgotPasswordRequest identity.
func (store *UserStore) FindUserByForgotPasswordRequest(req *pb.ForgotPasswordRequest) (*User, error) {
	var user User

	switch identity := req.GetIdentity().(type) {
	case *pb.ForgotPasswordRequest_UserId:
		result := store.DB.First(&user, "id = ?", identity.UserId)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find user by user ID: %v", result.Error)
		}
	case *pb.ForgotPasswordRequest_Email:
		result := store.DB.First(&user, "email = ?", identity.Email)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find user by email: %v", result.Error)
		}
	case *pb.ForgotPasswordRequest_Username:
		result := store.DB.First(&user, "username = ?", identity.Username)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to find user by username: %v", result.Error)
		}
	default:
		return nil, fmt.Errorf("unsupported identity type in request")
	}

	return &user, nil
}

// ReadUserPasswordByID retrieves the hashed password of a user by ID.
func (store *UserStore) ReadUserPasswordByID(userID string) (string, error) {
	var user User
	result := store.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		return "", fmt.Errorf("failed to get user password: %v", result.Error)
	}
	return user.HashedPassword, nil
}

// UpdateUserPassword updates a user's hashed password by user ID.
func (store *UserStore) UpdateUserPassword(userID, newPassword string) error {
	var user User
	result := store.DB.Model(&user).Where("id = ?", userID).Update("hashed_password", newPassword)
	if result.Error != nil {
		return fmt.Errorf("failed to update user password: %v", result.Error)
	}
	return nil
}

// ChangePassword hashes newPassword and persists it for the given user.
func (store *UserStore) ChangePassword(userID *uuid.UUID, newPassword string) (*User, error) {
	user, err := store.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}
	user.HashedPassword = utils.HASH_PASSWORD(newPassword)
	if _, err := store.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}
	return user, nil
}

// RetrieveAllUsers retrieves every user in the database.
func (store *UserStore) RetrieveAllUsers() ([]*User, error) {
	var users []*User
	result := store.DB.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all users: %v", result.Error)
	}
	return users, nil
}

// FindUsersByRoleIds retrieves all users assigned to any of the given role IDs.
// The return value is the raw proto representation to satisfy rbac.UserQuerier without
// causing an import cycle.
func (store *UserStore) FindUsersByRoleIds(roleIds []string) ([]*pb.User, error) {
	var users []*User
	err := store.DB.Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id IN ?", roleIds).
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by role IDs: %v", err)
	}
	pbUsers := make([]*pb.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, UserToPbUser(u))
	}
	return pbUsers, nil
}

// FindRolesByUserID retrieves all roles assigned to the given user.
// This satisfies rbac.UserQuerier without importing the user package from rbac.
func (store *UserStore) FindRolesByUserID(id *uuid.UUID) ([]*rbac.Role, error) {
	user, err := store.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}
	roles := make([]*rbac.Role, 0, len(user.Roles))
	for i := range user.Roles {
		roles = append(roles, &user.Roles[i])
	}
	return roles, nil
}

// AddRoleToUser appends a role to a user's role set.
func (store *UserStore) AddRoleToUser(userID, roleID string) (*User, error) {
	user := &User{}
	err := store.DB.Preload("Roles").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}
	role := &rbac.Role{}
	err = store.DB.First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find role by ID: %v", err)
	}
	err = store.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Association("Roles").Append(role)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add role to user: %v", err)
	}
	return user, nil
}

// RemoveRoleFromUser removes a specific role from a user's role set.
func (store *UserStore) RemoveRoleFromUser(userID, roleID string) (*User, error) {
	user := &User{}
	err := store.DB.Preload("Roles").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %v", err)
	}
	roles := make([]rbac.Role, 0)
	for _, r := range user.Roles {
		if r.ID.String() != roleID {
			roles = append(roles, r)
		}
	}
	err = store.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Association("Roles").Replace(roles)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove role from user: %v", err)
	}
	return user, nil
}

// AddUsersToRole assigns a role to multiple users by ID.
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

// DeactivateUser sets a user's status to INACTIVE.
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

// ActivateUser sets a user's status to ACTIVE.
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

// Search performs a fuzzy search over user fields with pagination.
func (store *UserStore) Search(query string, page int32, perPage int32) ([]*User, int32, error) {
	var users []*User
	var count int64
	offset := (page - 1) * perPage

	result := store.DB.Where(
		"(first_name ILIKE ? OR last_name ILIKE ? OR username ILIKE ? OR email ILIKE ? OR phone ILIKE ?)",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%",
	).Find(&users).Count(&count).Offset(int(offset)).Limit(int(perPage))
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to search users: %v", result.Error)
	}
	return users, int32(count), nil
}

// FindByFirebaseUID retrieves a user by Firebase UID.
func (store *UserStore) FindByFirebaseUID(firebaseUID string) (*User, error) {
	var user User
	result := store.DB.Preload(clause.Associations).First(&user, "firebase_uid = ?", firebaseUID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by Firebase UID: %v", result.Error)
	}
	return &user, nil
}

// Create creates a new user record.
func (store *UserStore) Create(user *User) (*User, error) {
	result := store.DB.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}
	return user, nil
}

// FindByIdsFromService retrieves users by a slice of ID strings, preloading associations.
func (store *UserStore) FindByIdsFromService(ids []string) ([]*User, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}
	var users []*User
	result := store.DB.Preload(clause.Associations).Where("id IN (?)", ids).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users by IDs: %v", result.Error)
	}
	return users, nil
}
