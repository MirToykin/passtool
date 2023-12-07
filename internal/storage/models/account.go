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

func (a *Account) FetchByLoginAndService(db *gorm.DB, login string, serviceID uint) error {
	return db.
		Preload("Password").
		Where("login = ? AND service_id = ?", login, serviceID).First(a).Error
}

func (a *Account) GetDecodedPassword(secret string, keyLen int) (string, error) {
	if a.Password.Encrypted == "" {
		return "", errors.New("account password is not loaded")
	}

	key := crypto.DeriveKey(secret, a.Password.Salt, keyLen)
	return crypto.Decrypt(key, a.Password.Encrypted)
}
