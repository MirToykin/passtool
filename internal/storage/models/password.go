package models

import (
	"errors"
	"github.com/MirToykin/passtool/internal/crypto"
	"gorm.io/gorm"
)

type Password struct {
	gorm.Model
	Encrypted string `gorm:"not null"`
	Salt      string `gorm:"not null"`
}

// GetDecrypted returns decoded password
func (p *Password) GetDecrypted(secret string, keyLen int) (string, error) {
	if p.Encrypted == "" {
		return "", errors.New("account password is not valid")
	}

	key := crypto.DeriveKey(secret, p.Salt, keyLen)
	return crypto.Decrypt(key, p.Encrypted)
}

func (p *Password) Save(db *gorm.DB) error {
	return db.Save(p).Error
}
