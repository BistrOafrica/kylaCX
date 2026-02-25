package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusStore struct {
	db *gorm.DB
}

func NewAgentStatusStore(db *gorm.DB) *StatusStore {
	return &StatusStore{
		db: db,
	}
}

func (store *StatusStore) Save(agentStatus *AgentStatus) error {
	if err := store.db.Save(agentStatus).Error; err != nil {
		return err
	}
	return nil
}

func (store *StatusStore) SaveStatusChange(agentStatus *AgentStatus, statusChange *StatusChange) error {
	if err := store.db.Save(statusChange).Error; err != nil {
		return err
	}

	agentStatus.StatusChanges = append(agentStatus.StatusChanges, *statusChange)
	if err := store.db.Save(agentStatus).Error; err != nil {
		return err
	}
	return nil
}

func (store *StatusStore) ReadAgentStatus(agentID uuid.UUID) (*AgentStatus, error) {
	var agentStatus AgentStatus
	if err := store.db.Where("agent_id = ?", agentID).Preload("StatusChanges").First(&agentStatus).Error; err != nil {
		return nil, err
	}
	return &agentStatus, nil
}

func (store *StatusStore) ReadLatestStatusChange(agentID uuid.UUID) (*StatusChange, error) {
	var statusChange StatusChange
	if err := store.db.Where("agent_id = ?", agentID).Order("created_at desc").First(&statusChange).Error; err != nil {
		return nil, err
	}
	return &statusChange, nil
}

func (store *StatusStore) ListStatusChanges(agentID uuid.UUID) ([]*StatusChange, error) {
	var statusChanges []*StatusChange
	if err := store.db.Where("agent_id = ?", agentID).Find(&statusChanges).Error; err != nil {
		return nil, err
	}
	return statusChanges, nil
}
