package models

import (
	"fmt"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Login      string `gorm:"index:idx_login_service,unique;not null"`
	ServiceID  uint   `gorm:"index:idx_login_service,unique;not null"`
	PasswordID uint   `gorm:"not null"`

	Service  Service
	Password Password
}

// FetchByLoginAndService fetches account with the given login for the given service
func (a *Account) FetchByLoginAndService(db *gorm.DB, login string, serviceID uint) error {
	return db.
		Preload("Password").
		Where("login = ? AND service_id = ?", login, serviceID).First(a).Error
}

// LoadPassword loads related password to account struct
func (a *Account) LoadPassword(db *gorm.DB) error {
	err := db.Model(Password{}).Where("id = ?", a.PasswordID).First(&a.Password).Error
	if err != nil {
		return fmt.Errorf("unable to load password: %w", err)
	}
	return nil
}

// List prepare query of all the accounts and return it
func (a *Account) List(db *gorm.DB) *gorm.DB {
	return db.Model(Account{})
}

// FindByLoginAndServiceID returns accounts query filtered by login and service id
func (a *Account) FindByLoginAndServiceID(db *gorm.DB, login string, serviceID uint) *gorm.DB {
	return a.List(db).Where("login = ? AND service_id = ?", login, serviceID)
}

// SaveWithPassword performs transactional save of password and account to database
func (a *Account) SaveWithPassword(db *gorm.DB, password *Password) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&password).Error; err != nil {
			return err
		}

		a.PasswordID = password.ID

		if err := tx.Create(&a).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to save account with password: %w", err)
	}

	return nil
}

// DeleteWithPassword performs transactional deletion of password and account from database
func (a *Account) DeleteWithPassword(db *gorm.DB) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(&a, a.ID).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Delete(&a.Password, a.PasswordID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to delete account with password: %w", err)
	}

	return nil
}
