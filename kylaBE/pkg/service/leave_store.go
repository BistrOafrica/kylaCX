package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaveStoreDB struct {
	DB *gorm.DB
}

func NewLeaveStore(db *gorm.DB) *LeaveStoreDB {
	return &LeaveStoreDB{
		DB: db,
	}
}

func (s *LeaveStoreDB) CreateLeaveType(leaveType *LeaveType) (*LeaveType, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// create leave type
		if err := tx.Create(leaveType).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return leaveType, nil
}

func (store *LeaveStoreDB) UpdateLeaveType(leaveType *LeaveType) (*LeaveType, error) {
	// Fields to exclude
	fieldsToExclude := []string{"Model", "ID", "CreatedAt", "DeletedAt", "CreatedBy", "OwnerID", "OwnerType"}
	// Read zero-valued fields
	fn := func(tx *gorm.DB) (*LeaveType, error) {
		tx.Begin()
		result := store.DB.Omit(fieldsToExclude...).Save(leaveType)
		if result.Error != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update leave type: %v", result.Error)
		}

		if len(leaveType.EarnedLeaveConditions) > 0 {
			// Delete existing conditions if any
			if err := tx.Unscoped().Where("leave_type_id = ?", leaveType.ID).
				Delete(&EarnedLeaveCondition{}).Error; err != nil {
				return leaveType, err
			}

			// Create new conditions
			for _, condition := range leaveType.EarnedLeaveConditions {
				if err := tx.Create(&condition).Error; err != nil {
					return leaveType, err
				}
			}
		}
		return leaveType, nil
	}
	return fn(store.DB)
}

func (store *LeaveStoreDB) DeleteEarnedLeaveConditions(leaveTypeId string) error {
	result := store.DB.Unscoped().Where("leave_type_id = ?", leaveTypeId).
		Delete(&EarnedLeaveCondition{})

	if result.Error == nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	if result.RowsAffected == 0 {
		return nil
	}

	return result.Error
}

func (store *LeaveStoreDB) ReadLeaveTypeById(md *RequestMetadata, id string) (*LeaveType, error) {
	var leaveType LeaveType
	result := store.DB.Preload(clause.Associations).First(&leaveType, "id = ?", id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find leave type by ID: %v", result.Error)
	}

	opScope := &OpScope{
		Owner: OwnerType(leaveType.OwnerType),
		ID:    leaveType.OwnerID.String(),
	}

	if !CheckOpScope(md, opScope) {
		return nil, status.Error(403, "Forbidden, You do not have access to access this resource")
	}

	return &leaveType, nil
}

func (store *LeaveStoreDB) ReadLeaveTypes(idsAllowingAccess []string, page int32, per_page int32) ([]*LeaveType, int32, error) {
	var leaveTypes []*LeaveType
	var count int64
	offset := (page - 1) * per_page

	query := store.DB.Where("owner_id IN (?)", idsAllowingAccess)
	result := query.Model(&LeaveType{}).Order("created_at desc")
	result.Count(&count)
	result = result.Preload(clause.Associations).Offset(int(offset)).Limit(int(per_page)).Find(&leaveTypes)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find leave types: %v", result.Error)
	}
	return leaveTypes, int32(count), nil
}

func (store *LeaveStoreDB) DeleteLeaveType(id string) error {
	var leaveType LeaveType
	result := store.DB.First(&leaveType, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to find leave type: %v", result.Error)
	}

	// Delete the conditions associated with the leave type
	store.DB.Where("leave_type_id = ?", id).Delete(&EarnedLeaveCondition{})

	// Delete the leave type
	result = store.DB.Delete(&leaveType)
	if result.Error != nil {
		return fmt.Errorf("failed to delete leave type: %v", result.Error)
	}

	return nil
}

func (store *LeaveStoreDB) CreateLeaveRequest(leaveRequest *LeaveRequest) (*LeaveRequest, error) {
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		// verify user can apply for this leave type
		var leaveType *LeaveType
		leaveTypeErr := store.DB.First(&leaveType, "id = ?", leaveRequest.LeaveTypeID).Error
		if leaveTypeErr != nil {
			return fmt.Errorf("leave type not found: %v", leaveTypeErr)
		}

		var user *User
		userErr := store.DB.First(&user, "id = ? and owner_id = ?", leaveRequest.UserID, leaveType.OwnerID).Error
		if userErr != nil {
			return fmt.Errorf("user not found or is ineligible: %v", userErr)
		}

		// create leave request
		if err := tx.Create(leaveRequest).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return leaveRequest, nil
}

func (store *LeaveStoreDB) ReadLeaveRequestById(md *RequestMetadata, id string) (*LeaveRequest, error) {
	var leaveReq LeaveRequest
	result := store.DB.Preload(clause.Associations).Preload("LeaveRequestEvents.User").First(&leaveReq, "id = ?", id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find leave request by ID: %v", result.Error)
	}

	opScope := &OpScope{
		Owner: OwnerType(leaveReq.LeaveType.OwnerType),
		ID:    leaveReq.LeaveType.OwnerID.String(),
	}

	if leaveReq.UserID != md.UserID && !CheckOpScope(md, opScope) {
		return nil, status.Error(403, "Forbidden, You do not have access to access this resource")
	}

	return &leaveReq, nil
}

func (store *LeaveStoreDB) ReadLeaveRequests(leaveTypeIDs []uuid.UUID, page int32, per_page int32) ([]*LeaveRequest, int32, error) {
	var leaveRequests []*LeaveRequest
	var count int64
	offset := (page - 1) * per_page

	query := store.DB.Where("leave_type_id IN (?)", leaveTypeIDs)
	result := query.Model(&LeaveRequest{}).Order("created_at desc")
	result.Count(&count)
	result = result.Preload(clause.Associations).Preload("LeaveRequestEvents.User").Offset(int(offset)).Limit(int(per_page)).Find(&leaveRequests)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find leave requests: %v", result.Error)
	}

	return leaveRequests, int32(count), nil
}

func (store *LeaveStoreDB) ReadMyLeaveRequests(userId uuid.UUID, page int32, per_page int32) ([]*LeaveRequest, int32, error) {
	var leaveRequests []*LeaveRequest
	var count int64
	offset := (page - 1) * per_page

	query := store.DB.Where("user_id = ?", userId.String())
	result := query.Model(&LeaveRequest{}).Order("created_at desc")
	result.Count(&count)
	result = result.Preload(clause.Associations).Preload("LeaveRequestEvents.User").Offset(int(offset)).Limit(int(per_page)).Find(&leaveRequests)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find leave requests for user %s: %v", userId, result.Error)
	}
	return leaveRequests, int32(count), nil
}

func (store *LeaveStoreDB) UpdateLeaveRequest(leaveRequest *LeaveRequest) (*LeaveRequest, error) {
	// Fields to exclude
	fieldsToExclude := []string{"Model", "ID", "CreatedAt", "DeletedAt", "CreatedBy", "UserID"}
	// Read zero-valued fields
	fn := func(tx *gorm.DB) (*LeaveRequest, error) {
		tx.Begin()
		result := store.DB.Omit(fieldsToExclude...).Save(leaveRequest).First(&leaveRequest, "id = ?", leaveRequest.ID)
		if result.Error != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update leave request: %v", result.Error)
		}
		return leaveRequest, nil
	}
	return fn(store.DB)
}

func (store *LeaveStoreDB) CreateLeaveRequestEvent(leaveRequestEvent *LeaveRequestEvent) (*LeaveRequestEvent, error) {
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		// create leave type
		if err := tx.Create(leaveRequestEvent).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return leaveRequestEvent, nil
}

func (store *LeaveStoreDB) UseLeaveBalanceUnits(userId uuid.UUID, leaveTypeId uuid.UUID, value int) error {
	return store.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&LeaveBalance{}).
			Where("user_id = ? AND leave_type_id = ?", userId, leaveTypeId).
			UpdateColumn("used", gorm.Expr("used + ?", value))

		if result.Error != nil {
			return fmt.Errorf("failed to update leave balance: %v", result.Error)
		}
		return nil
	})
}

func (store *LeaveStoreDB) RevertLeaveBalanceUnits(userId uuid.UUID, leaveTypeId uuid.UUID, value int) error {
	return store.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&LeaveBalance{}).
			Where("user_id = ? AND leave_type_id = ?", userId, leaveTypeId).
			UpdateColumn("used", gorm.Expr("used - ?", value))

		if result.Error != nil {
			return fmt.Errorf("failed to update leave balance: %v", result.Error)
		}
		return nil
	})
}

func (store *LeaveStoreDB) ReadLeaveBalance(userId uuid.UUID, leaveTypeId uuid.UUID) (*LeaveBalance, error) {
	var leaveBalance *LeaveBalance
	err := store.DB.Where("user_id = ? AND leave_type_id = ?", userId, leaveTypeId).Preload(clause.Associations).First(&leaveBalance).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// if record does not exist, create a new one if user is eligible
			var leaveType *LeaveType
			leaveTypeErr := store.DB.First(&leaveType, "id = ?", leaveTypeId).Error
			if leaveTypeErr != nil {
				return nil, fmt.Errorf("leave type not found: %v", leaveTypeErr)
			}

			var user *User
			userErr := store.DB.First(&user, "id = ? and owner_id = ?", userId, leaveType.OwnerID).Error
			if userErr != nil {
				return nil, fmt.Errorf("user not found or is ineligible: %v", userErr)
			}

			newLeaveBalance := &LeaveBalance{
				ID:            uuid.New(),
				UserID:        userId,
				LeaveTypeID:   leaveTypeId,
				TotalEligible: leaveType.TimeLimit,
				Used:          0,
				BalanceUnit:   leaveType.TimeLimitUnit,
			}
			result := store.DB.Save(newLeaveBalance)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to save leave balance: %v", result.Error)
			}

			newLeaveBalance.LeaveType = *leaveType
			newLeaveBalance.User = *user

			return newLeaveBalance, nil
		}
	}

	return leaveBalance, err
}

func (store *LeaveStoreDB) ReadLeaveRequestsMetrics(leaveTypeIDs []uuid.UUID) ([]*LeaveRequestsMetrics, error) {
	var metrics []*LeaveRequestsMetrics

	if len(leaveTypeIDs) == 0 {
		return metrics, nil
	}

	err := store.DB.Model(&LeaveRequest{}).
		Select("status, count(*) as total_count").
		Where("leave_type_id IN (?)", leaveTypeIDs).
		Group("status").
		Scan(&metrics).
		Error

	if err != nil {
		return nil, err
	}

	return metrics, nil
}
