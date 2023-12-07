package models

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"`

	Accounts []Account
}

func (s *Service) FetchByName(db *gorm.DB, name string, withAccounts bool) error {
	if withAccounts {
		db = db.Preload("Accounts")
	}
	return db.First(s, "name", name).Error
}
