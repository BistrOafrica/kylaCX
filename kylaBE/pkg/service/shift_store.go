package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShiftStore struct {
	DB *gorm.DB
}

func NewShiftStore(db *gorm.DB) *ShiftStore {
	return &ShiftStore{
		DB: db,
	}
}

func (s *ShiftStore) SaveShift(shift *Shift) (*Shift, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// create shift
		if err := tx.Create(shift).Error; err != nil {
			return err
		}

		// add schedules associations
		for _, schedule := range shift.ShiftSchedules {
			if err := tx.Model(&shift).Association("ShiftSchedules").Append(schedule); err != nil {
				return err
			}

			// add breaks associations
			for _, breakItem := range schedule.ScheduleBreaks {
				if err := tx.Model(&schedule).Association("ShiftBreaks").Append(breakItem); err != nil {
					return err
				}

				// add agent status change association
				if breakItem.AgentStatusChange != (StatusChange{}) {
					if err := tx.Model(&breakItem).Association("StatusChanges").Append(breakItem.AgentStatusChange); err != nil {
						return err
					}
				}
			}
		}

		// add users associations
		for _, user := range shift.ShiftUsers {
			if err := tx.Model(&shift).Association("Users").Append(user); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return shift, nil
}

func (s *ShiftStore) ReadShift(id uuid.UUID, opScope *OpScope) (*Shift, error) {
	shift := &Shift{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("ShiftSchedules").Preload("ShiftUsers").Preload("ShiftSchedules.Breaks").Preload("ShiftSchedules.Breaks.AgentStatusChange").First(shift, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return shift, nil
}

func (s *ShiftStore) ReadShifts(scopeIds []string) ([]*Shift, error) {
	shifts := []*Shift{}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Preload("ShiftSchedules").
			Preload("ShiftUsers").
			Preload("ShiftSchedules.Breaks").
			Preload("ShiftSchedules.Breaks.AgentStatusChange").
			Where("owner_id IN ? ", scopeIds).
			Find(&shifts).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return shifts, nil
}

func (s *ShiftStore) UpdateShift(shift *Shift) (*Shift, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// update shift
		if err := tx.Omit(
			"OwnerID", "OwnerType", "CreatedAt",
			"ID", "SerialNumber",
		).Save(shift).Error; err != nil {
			return err
		}

		// add schedules associations
		for _, schedule := range shift.ShiftSchedules {
			if err := tx.Model(&shift).Association("Schedules").Append(schedule); err != nil {
				return err
			}

			// add breaks associations
			for _, breakItem := range schedule.ScheduleBreaks {
				if err := tx.Model(&schedule).Association("Breaks").Append(breakItem); err != nil {
					return err
				}

				// add agent status change association
				if breakItem.AgentStatusChange != (StatusChange{}) {
					if err := tx.Model(&breakItem).Association("AgentStatusChange").Append(breakItem.AgentStatusChange); err != nil {
						return err
					}
				}
			}
		}

		// add users associations
		for _, user := range shift.ShiftUsers {
			if err := tx.Model(&shift).Association("Users").Append(user); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return shift, nil
}

func (s *ShiftStore) DeleteShift(id uuid.UUID) error {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// find the shift
		shift := &Shift{}
		if err := tx.Preload("Schedules.Breaks.AgentStatusChange").First(shift, id).Error; err != nil {
			return err
		}

		// delete associated schedules, breaks, and status changes
		for _, schedule := range shift.ShiftSchedules {
			for _, breakItem := range schedule.ScheduleBreaks {
				if err := tx.Delete(&breakItem.AgentStatusChange).Error; err != nil {
					return err
				}
				if err := tx.Delete(&breakItem).Error; err != nil {
					return err
				}
			}
			if err := tx.Delete(&schedule).Error; err != nil {
				return err
			}
		}

		// delete shift
		if err := tx.Delete(&Shift{}, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *ShiftStore) AssignUsersToShift(shiftID uuid.UUID, userIDs []uuid.UUID) error {
	shift := &Shift{}
	err := s.DB.First(shift, shiftID).Error
	if err != nil {
		return err
	}

	users := make([]User, len(userIDs))
	for i, userID := range userIDs {
		users[i] = User{ID: userID}
	}

	err = s.DB.Model(shift).Association("Users").Append(users)
	if err != nil {
		return err
	}
	return nil
}

func (s *ShiftStore) RemoveUsersFromShift(shiftID uuid.UUID, userIDs []uuid.UUID) error {
	shift := &Shift{}
	err := s.DB.First(shift, shiftID).Error
	if err != nil {
		return err
	}

	users := make([]User, len(userIDs))
	for i, userID := range userIDs {
		users[i] = User{ID: userID}
	}

	err = s.DB.Model(shift).Association("Users").Delete(users)
	if err != nil {
		return err
	}
	return nil
}

func (s *ShiftStore) ListShiftUsers(shiftID uuid.UUID) ([]User, error) {
	shift := &Shift{}
	err := s.DB.Preload("Users").First(shift, shiftID).Error
	if err != nil {
		return nil, err
	}
	return shift.ShiftUsers, nil
}
