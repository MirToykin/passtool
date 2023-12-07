package models

import (
	"errors"
	"github.com/MirToykin/passtool/internal/crypto"
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
	return db.Model(Password{}).Where("id = ?", a.PasswordID).First(&a.Password).Error
}

// GetDecodedPassword returns decoded account password
func (a *Account) GetDecodedPassword(secret string, keyLen int) (string, error) {
	if a.Password.Encrypted == "" {
		return "", errors.New("account password is not loaded")
	}

	key := crypto.DeriveKey(secret, a.Password.Salt, keyLen)
	return crypto.Decrypt(key, a.Password.Encrypted)
}
