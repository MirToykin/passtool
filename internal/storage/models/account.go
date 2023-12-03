package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	Login      string `gorm:"index:idx_login_service,unique;not null"`
	ServiceID  uint   `gorm:"index:idx_login_service,unique;not null"`
	PasswordID uint   `gorm:"not null"`

	Service  Service
	Password Password
}
