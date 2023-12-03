package models

import "gorm.io/gorm"

type Password struct {
	gorm.Model
	Encrypted string `gorm:"not null"`
}
