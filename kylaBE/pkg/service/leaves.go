package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemunerationType string

const (
	PAID      RemunerationType = "PAID"
	UNPAID    RemunerationType = "UNPAID"
	HALF_PAID RemunerationType = "HALF_PAID"
)

type EarnedType string

const (
	UPFRONT EarnedType = "UPFRONT"
	EARNED  EarnedType = "EARNED"
)

type TimeUnit string

const (
	DAY  TimeUnit = "DAY"
	HOUR TimeUnit = "HOUR"
)

type TimeLimitCycle string

const (
	ANNUALLY         TimeLimitCycle = "ANNUALLY"
	SEMI_ANNUALLY    TimeLimitCycle = "SEMI_ANNUALLY"
	QUARTER_ANNUALLY TimeLimitCycle = "QUARTER_ANNUALLY"
	MONTHLY          TimeLimitCycle = "MONTHLY"
	PER_WEEK         TimeLimitCycle = "PER_WEEK"
)

type AccrualPolicy string

const (
	CARRY_FORWARD AccrualPolicy = "CARRY_FORWARD"
	EXPIRE        AccrualPolicy = "EXPIRE"
)

type LeaveStatus string

const (
	LEAVE_PENDING   LeaveStatus = "LEAVE_PENDING"
	LEAVE_APPROVED  LeaveStatus = "LEAVE_APPROVED"
	LEAVE_REJECTED  LeaveStatus = "LEAVE_REJECTED"
	LEAVE_ONGOING   LeaveStatus = "LEAVE_ONGOING"
	LEAVE_APPEALED  LeaveStatus = "LEAVE_APPEALED"
	LEAVE_CANCELLED LeaveStatus = "LEAVE_CANCELLED"
	LEAVE_ENDED     LeaveStatus = "LEAVE_ENDED"
)

type LeaveAction string

const (
	APPROVED  LeaveAction = "APPROVED"
	REJECTED  LeaveAction = "REJECTED"
	CANCELLED LeaveAction = "CANCELLED"
	APPEALED  LeaveAction = "APPEALED"
	STARTED   LeaveAction = "STARTED"
	ENDED     LeaveAction = "ENDED"
)

type LeaveType struct {
	gorm.Model
	ID                     uuid.UUID              `gorm:"primarykey;type:uuid;not null;index"`
	Name                   string                 `gorm:"type:varchar(255);index"`
	Description            string                 `gorm:"type:varchar(255);"`
	RemunerationType       RemunerationType       `gorm:"not null; default:PAID"`
	EarnedType             EarnedType             `gorm:"not null; default:UPFRONT"`
	EarnedLeaveConditions  []EarnedLeaveCondition `gorm:"foreignKey:LeaveTypeID;constraint:OnDelete:CASCADE"`
	TimeLimit              int                    `gorm:"type:integer;"`
	TimeLimitUnit          TimeUnit               `gorm:"default:null;"`
	TimeLimitCycle         TimeLimitCycle         `gorm:"default:null;"`
	ApplyToWorkingTimeOnly bool                   `gorm:"type:boolean; default:true;"`
	AccrualPolicy          AccrualPolicy          `gorm:"default:EXPIRE;"`
	CreatedBy              string                 `gorm:"column:created_by; default:null"`
	OwnerID                uuid.UUID              `gorm:"not null; type:uuid;index"`
	OwnerType              OwnerType              `gorm:"not null; default:ORGANISATIONS;index"`
}

type EarnedLeaveCondition struct {
	gorm.Model
	ID              uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	LeaveTypeID     uuid.UUID `gorm:"type:uuid;not null;index"`
	LeaveType       LeaveType `gorm:"constraint:OnDelete:CASCADE"`
	WorkingTime     int       `gorm:"type:integer;"`
	WorkingTimeUnit TimeUnit  `gorm:"default:null;"`
	RewardTime      int       `gorm:"type:integer;"`
	RewardTimeUnit  TimeUnit  `gorm:"default:null;"`
}

type LeaveBalance struct {
	gorm.Model
	ID            uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	UserID        uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_userid_leavetypeid"`
	User          User      `gorm:"constraint:OnDelete:CASCADE"`
	LeaveTypeID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_userid_leavetypeid"`
	LeaveType     LeaveType `gorm:"constraint:OnDelete:CASCADE"`
	TotalEligible int       `gorm:"type:integer;default:0"`
	Used          int       `gorm:"type:integer;default:0"`
	Remaining     int       `gorm:"->;generated always as (total_eligible - used) stored"` // Auto-calculated
	BalanceUnit   TimeUnit  `gorm:"default:null;"`
}

type LeaveRequestEvent struct {
	gorm.Model
	ID             uuid.UUID    `gorm:"primarykey;type:uuid;not null"`
	LeaveRequestID uuid.UUID    `gorm:"type:uuid;not null"`
	LeaveRequest   LeaveRequest `gorm:"constraint:OnDelete:CASCADE"`
	Action         LeaveAction  `gorm:"type:varchar(255);not null"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null"`
	User           User         `gorm:"constraint:OnDelete:CASCADE"`
}

type LeaveRequest struct {
	gorm.Model
	ID                      uuid.UUID                `gorm:"primarykey;type:uuid;not null"`
	UserID                  uuid.UUID                `gorm:"type:uuid;not null"`
	User                    User                     `gorm:"constraint:OnDelete:CASCADE"`
	LeaveTypeID             uuid.UUID                `gorm:"type:uuid;not null"`
	LeaveType               LeaveType                `gorm:"constraint:OnDelete:CASCADE"`
	StartDate               string                   `gorm:"type:varchar(255);not null"`
	EndDate                 string                   `gorm:"type:varchar(255);not null"`
	StartTime               string                   `gorm:"type:varchar(255);"`
	EndTime                 string                   `gorm:"type:varchar(255);"`
	Duration                int                      `gorm:"type:integer;"`
	DurationUnit            TimeUnit                 `gorm:"default:null;"`
	ApplierComment          string                   `gorm:"type:text;"`
	Status                  LeaveStatus              `gorm:"default:LEAVE_PENDING;"`
	LeaveRequestAttachments []LeaveRequestAttachment `gorm:"foreignKey:LeaveRequestID;constraint:OnDelete:CASCADE"`
	ApproverID              string                   `gorm:"default:null"`
	ApproverComment         string                   `gorm:"type:text;"`
	IsRetrospective         bool                     `gorm:"type:boolean; default:false;"`
	AppealReason            string                   `gorm:"type:text;"`
	CreatedBy               string                   `gorm:"column:created_by; default:null"`
	LeaveRequestEvents      []LeaveRequestEvent      `gorm:"foreignKey:LeaveRequestID;constraint:OnDelete:CASCADE"`
}

type LeaveRequestAttachment struct {
	gorm.Model
	ID             uuid.UUID    `gorm:"primarykey;type:uuid;not null"`
	LeaveRequestID uuid.UUID    `gorm:"type:uuid;not null"`
	LeaveRequest   LeaveRequest `gorm:"constraint:OnDelete:CASCADE"`
	FileName       string       `gorm:"type:text"`
	FileUrl        string       `gorm:"type:text;"`
}

type LeaveRequestsMetrics struct {
	TotalCount int32
	Status     LeaveStatus
}
