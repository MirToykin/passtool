package sqlite

import (
	"fmt"
	"github.com/MirToykin/passtool/internal/storage/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(storagePath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(storagePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	err = db.AutoMigrate(&models.Service{}, &models.Account{}, models.Password{})
	if err != nil {
		return nil, fmt.Errorf("failed apply migrations: %w", err)
	}

	return db, nil
}
