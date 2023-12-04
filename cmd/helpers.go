package cmd

import (
	"errors"
	"github.com/MirToykin/passtool/internal/crypto"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	passGenerator "github.com/sethvargo/go-password/password"
	"gorm.io/gorm"
	"log"
)

func fetchOrCreateService(service *models.Service, serviceName string) {
	err := db.First(&service, "name", serviceName).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		service.Name = serviceName
		err = db.Create(&service).Error
	}

	checkSimpleError(err, "unable to create service")
}

func getAccountsCount(account *models.Account, login string, serviceID uint) int64 {
	var count int64
	result := db.Model(&account).Where("login = ? AND service_id = ?", login, serviceID).Count(&count)
	checkSimpleError(result.Error, "unable to check account existence")
	return count
}

func getPassPhrase() string {
	secretKey, err := cli.GetSensitiveUserInput("Enter secret pass phrase: ")
	checkSimpleError(err, "unable to get passphrase")
	return secretKey
}

func saveAccountWithPassword(account *models.Account, password *models.Password) {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&password).Error; err != nil {
			return err
		}

		account.PasswordID = password.ID

		if err := tx.Create(&account).Error; err != nil {
			return err
		}

		return nil
	})
	checkSimpleError(err, "unable to create password")
}

func checkSimpleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func encryptPassword(password *models.Password, userPassword, passPhrase string) {
	// TODO вынести параметры генерации salt в конфиг
	salt, err := passGenerator.Generate(64, 10, 10, false, false)
	checkSimpleError(err, "unable to get salt")

	keyLen := 32 // TODO вынести в конфиг
	key := crypto.DeriveKey(passPhrase, salt, keyLen)

	encryptedPassword, err := crypto.Encrypt(key, userPassword)
	checkSimpleError(err, "unable to encrypt the password")

	password.Encrypted = encryptedPassword
	password.Salt = salt
}