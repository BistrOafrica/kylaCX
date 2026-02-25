package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppStore struct {
	db *gorm.DB
}

func NewAPIAppStore(db *gorm.DB) *AppStore {
	return &AppStore{
		db: db,
	}
}

// Apps
func (store *AppStore) SaveApp(app *App) (*App, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(app).Error
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (store *AppStore) FindAppByID(id string, scope *OpScope) (*App, error) {
	var app App
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(&app, "id = ? AND owner_type = ? AND owner_id = ?", id, scope.Owner, scope.ID).Error
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (store *AppStore) FindAppByName(name string, scope *OpScope) (*App, error) {
	var app App
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(&app, "name = ? AND owner_type = ? AND owner_id = ?", name, scope.Owner, scope.ID).Error

	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (store *AppStore) FindAllAppsByOrganisationID(organisationID uuid.UUID) ([]*App, error) {
	var apps []*App
	result := store.db.Find(&apps, "organisation_id = ?", organisationID)
	if result.Error != nil {
		return nil, result.Error
	}
	return apps, nil
}

func (store *AppStore) FindallTemplates() ([]*App, error) {
	var apps []*App
	result := store.db.Find(&apps, "is_template = ?", true)
	if result.Error != nil {
		return nil, result.Error
	}
	return apps, nil
}

func (store *AppStore) DeleteApp(appID string) error {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", appID).Delete(&App{}).Error
	})
	if err != nil {
		return err
	}
	return nil
}

func (store *AppStore) UpdateApp(app *App) (*App, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Omit("Secret", "Token", "ID", "BranchID", "OrganisationID", "OwnerID", "OwnerType").Save(app).Error
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (store *AppStore) UpdateAppStatus(appID string, status string) (*App, error) {
	var app App
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&app, "id = ?", appID).Error; err != nil {
			return err
		}
		app.Status = status
		return tx.Save(&app).Error
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (store *AppStore) FindByToken(token string) (*App, error) {
	var app App
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(&app, "token = ?", token).Error
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (store *AppStore) UpdateTokenAndSecret(newApp *App) (*App, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&App{}).Where("id = ?", newApp.ID).Updates(App{Token: newApp.Token, Secret: newApp.Secret}).Error
	})
	if err != nil {
		return nil, err
	}
	return newApp, nil
}

// ReadConsoleApps: Read all console apps
func (store *AppStore) FindAll() ([]*App, error) {
	var apps []*App
	err := store.db.Transaction(func(tx *gorm.DB) error {
		return tx.Find(&apps).Error
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}
