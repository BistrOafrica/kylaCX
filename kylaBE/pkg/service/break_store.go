package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BreakStore interface {
	SaveBreak(*ScheduleBreak) (*ScheduleBreak, error)
	ReadBreak(id uuid.UUID, opScope *OpScope) (*ScheduleBreak, error)
	ListBreaks(scheduleID uuid.UUID, opScope *OpScope) ([]*ScheduleBreak, error)
	UpdateBreak(*ScheduleBreak) (*ScheduleBreak, error)
	DeleteBreak(uuid.UUID) error
}

type BreakStoreDB struct {
	DB *gorm.DB
}

func NewBreakStore(db *gorm.DB) *BreakStoreDB {
	return &BreakStoreDB{
		DB: db,
	}
}

func (s *BreakStoreDB) SaveBreak(breakItem *ScheduleBreak) (*ScheduleBreak, error) {
	err := s.DB.Create(breakItem).Error
	if err != nil {
		return nil, err
	}
	return breakItem, nil
}

func (s *BreakStoreDB) ReadBreak(id uuid.UUID, opScope *OpScope) (*ScheduleBreak, error) {
	breakItem := &ScheduleBreak{}
	err := s.DB.First(breakItem, id).Error
	if err != nil {
		return nil, err
	}
	return breakItem, nil
}

func (s *BreakStoreDB) ListBreaks(scheduleID uuid.UUID, opScope *OpScope) ([]*ScheduleBreak, error) {
	breaks := []*ScheduleBreak{}
	err := s.DB.Where("schedule_id = ?", scheduleID).Find(&breaks).Error
	if err != nil {
		return nil, err
	}
	return breaks, nil
}

func (s *BreakStoreDB) UpdateBreak(breakItem *ScheduleBreak) (*ScheduleBreak, error) {
	err := s.DB.Omit(
		"OwnerID", "OwnerType", "CreatedAt",
		"ID", "SerialNumber",
	).Save(breakItem).Error
	if err != nil {
		return nil, err
	}
	return breakItem, nil
}

func (s *BreakStoreDB) DeleteBreak(id uuid.UUID) error {
	err := s.DB.Delete(&ScheduleBreak{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
