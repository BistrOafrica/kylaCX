package service

import "gorm.io/gorm"

type OnboardingStore struct {
	db *gorm.DB
}

func NewOnboardingStore(db *gorm.DB) *OnboardingStore {
	return &OnboardingStore{db: db}
}

func (store *OnboardingStore) CreateOnboarding(onboarding *Onboarding) (*Onboarding, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(onboarding).First(&onboarding).Error; err != nil {
			return err
		}
		return nil
	})
	return onboarding, err
}

func (store *OnboardingStore) GetOnboardingByID(id string) (*Onboarding, error) {
	var onboarding Onboarding
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&onboarding).Error; err != nil {
			return err
		}
		return nil
	})
	return &onboarding, err
}

func (store *OnboardingStore) UpdateOnboarding(onboarding *Onboarding) (*Onboarding, error) {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(onboarding).First(&onboarding).Error; err != nil {
			return err
		}
		return nil
	})
	return onboarding, err
}

func (store *OnboardingStore) DeleteOnboarding(id string) error {
	err := store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&Onboarding{}).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

type ListOnboardingsParams struct {
	Page     int
	PageSize int
	Status   string
	SortBy   string
	SortDesc bool
}

func (store *OnboardingStore) ListOnboardings(params ListOnboardingsParams) ([]Onboarding, int64, error) {
	var onboardings []Onboarding
	var total int64

	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}

	offset := (params.Page - 1) * params.PageSize

	err := store.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&Onboarding{})

		// Apply filters
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}

		// Get total count
		if err := query.Count(&total).Error; err != nil {
			return err
		}

		// Apply sorting
		sortOrder := "asc"
		if params.SortDesc {
			sortOrder = "desc"
		}
		query = query.Order(params.SortBy + " " + sortOrder)

		// Apply pagination
		if err := query.Offset(offset).Limit(params.PageSize).Find(&onboardings).Error; err != nil {
			return err
		}

		return nil
	})

	return onboardings, total, err
}
