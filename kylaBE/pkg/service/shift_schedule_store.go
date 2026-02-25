package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShiftScheduleStore struct {
	DB *gorm.DB
}

func NewShiftScheduleStore(db *gorm.DB) *ShiftScheduleStore {
	return &ShiftScheduleStore{
		DB: db,
	}
}

func (s *ShiftScheduleStore) SaveSchedule(schedule *ShiftSchedule) (*ShiftSchedule, error) {
	err := s.DB.Create(schedule).Error
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *ShiftScheduleStore) ReadSchedule(id uuid.UUID, opScope *OpScope) (*ShiftSchedule, error) {
	schedule := &ShiftSchedule{}
	err := s.DB.First(schedule, id).Error
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *ShiftScheduleStore) ListSchedules(shiftID uuid.UUID, opScope *OpScope) ([]*ShiftSchedule, error) {
	schedules := []*ShiftSchedule{}
	err := s.DB.Where("shift_id = ?", shiftID).Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (s *ShiftScheduleStore) UpdateSchedule(schedule *ShiftSchedule) (*ShiftSchedule, error) {
	err := s.DB.Omit(
		"OwnerID", "OwnerType", "CreatedAt",
		"ID", "SerialNumber",
	).Save(schedule).Error
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *ShiftScheduleStore) DeleteSchedule(id uuid.UUID) error {
	err := s.DB.Delete(&ShiftSchedule{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

// clock in and out
func (s *ShiftScheduleStore) ClockIn(record *UserShiftRecord) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ShiftScheduleStore) ClockOut(record *UserShiftRecord) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ShiftScheduleStore) TakeBreak(record *BreakRecord) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ShiftScheduleStore) ResumeBreak(record *BreakRecord) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
