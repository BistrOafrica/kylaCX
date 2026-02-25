package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shift struct {
	gorm.Model
	ID                       uuid.UUID       `gorm:"primarykey;type:uuid;not null"`
	SerialNumber             string          `gorm:"type:varchar(255);"`
	Name                     string          `gorm:"type:varchar(255);"`
	Description              string          `gorm:"type:varchar(255);"`
	StartTime                string          `gorm:"type:varchar(255);"`
	EndTime                  string          `gorm:"type:varchar(255);"`
	ShiftType                string          `gorm:"type:varchar(255);"`
	ShiftPeriod              ShiftPeriod     `gorm:"type:integer;"`
	ShiftStatus              string          `gorm:"type:varchar(255);"`
	MinUsersPerShift         int             `gorm:"type:integer;"`
	MaxUsersPerShift         int             `gorm:"type:integer;"`
	OverTimeAllowed          bool            `gorm:"type:boolean;"`
	OverTimeLimit            int             `gorm:"type:integer;"`
	OverTimeLimitUnit        string          `gorm:"type:varchar(255);"`
	SchedulingType           SchedulingType  `gorm:"type:integer;"`
	HourlyRate               float32         `gorm:"type:float;"`
	HourlyRateCurrency       string          `gorm:"type:varchar(255);"`
	HourlyRateCurrencySymbol CurrencySymbol  `gorm:"type:integer;"`
	Color                    string          `gorm:"type:varchar(255);"`
	Banner                   string          `gorm:"type:varchar(255);"`
	OwnerType                OwnerType       `gorm:"type:text;"`
	OwnerID                  uuid.UUID       `gorm:"type:uuid;"`
	CreatedBy                string          `gorm:"type:varchar(255);"`
	UpdatedBy                string          `gorm:"type:varchar(255);"`
	ShiftUsers               []User          `gorm:"many2many:shift_users;"`
	ShiftSchedules           []ShiftSchedule `gorm:"foreignKey:ShiftID"`
}

type ShiftPeriod int

const (
	UNKNOWN_SHIFT_TYPE ShiftPeriod = iota
	PERMANENT
	TEMPORARY
)

type TimeUnits int

const (
	UNKNOWN_OVERTIME_LIMIT_UNIT TimeUnits = iota
	MINUTES
	HOURS
	DAYS
	WEEKS
	MONTHS
	YEARS
)

type SchedulingType int

const (
	UNKNOWN_SCHEDULING_TYPE SchedulingType = iota
	DAILY_SCHEDULING
	WEEKLY_SCHEDULING
	MONTHLY_SCHEDULING
	QUARTERLY_SCHEDULING
	YEARLY_SCHEDULING
)

type CurrencySymbol int

const (
	UNKNOWN_CURRENCY_SYMBOL CurrencySymbol = iota
	USD                                    = "$"
	EUR                                    = "€"
	GBP                                    = "£"
	NGN                                    = "₦"
	KES                                    = "Ksh"
	UGX                                    = "Ush"
	ZAR                                    = "R"
	TSH                                    = "Tsh"
)

type ShiftScheduleType int

const (
	WEEKLY_SCHEDULE ShiftScheduleType = iota
	MONTHLY_SCHEDULE
	MONDAY_SCHEDULE
	TUESDAY_SCHEDULE
	WEDNESDAY_SCHEDULE
	THURSDAY_SCHEDULE
	FRIDAY_SCHEDULE
	SATURDAY_SCHEDULE
	SUNDAY_SCHEDULE
)

type ShiftSchedule struct {
	gorm.Model
	ID                uuid.UUID       `gorm:"primarykey;type:uuid;not null"`
	SerialNumber      string          `gorm:"type:varchar(255);"`
	ShiftID           uuid.UUID       `gorm:"type:uuid;"`
	StartDate         string          `gorm:"type:varchar(255);"`
	EndDate           string          `gorm:"type:varchar(255);"`
	StartTime         string          `gorm:"type:varchar(255);"`
	EndTime           string          `gorm:"type:varchar(255);"`
	MinUsers          int             `gorm:"type:integer;"`
	MaxUsers          int             `gorm:"type:integer;"`
	ShiftStatus       string          `gorm:"type:varchar(255);"`
	OwnerType         OwnerType       `gorm:"type:text;"`
	OwnerID           uuid.UUID       `gorm:"type:uuid;"`
	CreatedBy         string          `gorm:"type:varchar(255);"`
	UpdatedBy         string          `gorm:"type:varchar(255);"`
	ScheduleBreaks    []ScheduleBreak `gorm:"foreignKey:ShiftScheduleID"`
	ShiftScheduleType ShiftScheduleType
}

type ScheduleBreak struct {
	gorm.Model
	ID                  uuid.UUID         `gorm:"primarykey;type:uuid;not null"`
	SerialNumber        string            `gorm:"type:varchar(255);"`
	ShiftScheduleID     uuid.UUID         `gorm:"type:uuid;"`
	StartTime           string            `gorm:"type:varchar(255);"`
	EndTime             string            `gorm:"type:varchar(255);"`
	BreakPeriod         string            `gorm:"type:varchar(255);"`
	BreakPeriodUnit     TimeUnits         `gorm:"type:integer;"`
	ShiftScheduleType   ShiftScheduleType `gorm:"type:integer;"`
	MinUsersRequired    int               `gorm:"type:integer;"`
	AgentStatusID       uuid.UUID         `gorm:"type:uuid;"`
	AgentStatusChangeID uuid.UUID         `gorm:"type:uuid;"`
	AgentStatusChange   StatusChange
	OwnerType           OwnerType `gorm:"type:text; not null;"`
	OwnerID             uuid.UUID `gorm:"type:uuid; not null;"`
	Partial             bool      `gorm:"type:boolean;"`
	Mandatory           bool      `gorm:"type:boolean;"`
	CreatedBy           string    `gorm:"type:varchar(255);"`
	UpdatedBy           string    `gorm:"type:varchar(255);"`
}

// User Take Break
type BreakRecord struct {
	gorm.Model
	ID                uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	UserID            uuid.UUID `gorm:"type:uuid;not null"`
	BreakID           uuid.UUID `gorm:"type:uuid;not null"`
	UserShiftRecordID uuid.UUID `gorm:"type:uuid;not null"`
	StartTime         string    `gorm:"type:varchar(255);not null"`
	EndTime           string    `gorm:"type:varchar(255);not null"`
	CreatedBy         string    `gorm:"type:varchar(255);"`
	UpdatedBy         string    `gorm:"type:varchar(255);"`
}

type UserShiftRecord struct {
	gorm.Model
	ID        uuid.UUID     `gorm:"primarykey;type:uuid;not null"`
	UserID    uuid.UUID     `gorm:"type:uuid;not null"`
	ShiftID   uuid.UUID     `gorm:"type:uuid;not null"`
	StartTime string        `gorm:"type:varchar(255);not null"`
	EndTime   string        `gorm:"type:varchar(255);not null"`
	CreatedBy string        `gorm:"type:varchar(255);"`
	UpdatedBy string        `gorm:"type:varchar(255);"`
	Breaks    []BreakRecord `gorm:"foreignKey:UserShiftRecordID"`
}
