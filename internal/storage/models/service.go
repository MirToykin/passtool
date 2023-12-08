package models

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"`

	Accounts []Account
}

// FetchByName fetches service by its name.
// withAccounts defines whether it needs to prefetch service related accounts or not.
func (s *Service) FetchByName(db *gorm.DB, name string, withAccounts bool) error {
	if withAccounts {
		db = db.Preload("Accounts")
	}
	return db.First(s, "name", name).Error
}

// List prepare query of all the services and return it
func (s *Service) List(db *gorm.DB) *gorm.DB {
	return db.Model(Service{})
}

// GetList fetches all the services and return them
func (s *Service) GetList(db *gorm.DB, withAccounts bool) ([]Service, error) {
	var services []Service

	db = s.List(db)
	if withAccounts {
		db = db.Preload("Accounts")
	}

	err := db.Find(&services).Error

	return services, err
}

// GetAccountsMap returns map of accounts where keys are their serial starting numbers from 1
func (s *Service) GetAccountsMap() map[int]Account {
	aMap := make(map[int]Account)
	for i, account := range s.Accounts {
		aMap[i+1] = account
	}

	return aMap
}
