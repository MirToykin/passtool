package cmd

import (
	"errors"
	"fmt"
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/crypto"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	passGenerator "github.com/sethvargo/go-password/password"
	"gorm.io/gorm"
	"log"
)

// fetchOrCreateService fetches existing or creates new Service instance
func fetchOrCreateService(service *models.Service, serviceName string) {
	err := db.First(&service, "name", serviceName).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		service.Name = serviceName
		err = db.Create(&service).Error
	}

	checkSimpleError(err, "unable to create service")
}

// getAccountsCount returns number of accounts with a given login for a given service
func getAccountsCount(account *models.Account, login string, serviceID uint) int64 {
	var count int64
	result := db.Model(&account).Where("login = ? AND service_id = ?", login, serviceID).Count(&count)
	checkSimpleError(result.Error, "unable to check account existence")
	return count
}

// getPassPhrase returns pass phrase given by user (handle possible errors)
func getPassPhrase(confirm bool) string {
	postfix := ""
	if confirm {
		postfix = " again"
	}

	secretKey, err := cli.GetSensitiveUserInput(fmt.Sprintf("Enter secret pass phrase%s: ", postfix))
	checkSimpleError(err, "unable to get passphrase")
	return secretKey
}

// getPassPhraseWithConfirmation handles getting pass phrase with confirmation
func getPassPhraseWithConfirmation() string {
	for {
		pass1 := getPassPhrase(false)
		pass2 := getPassPhrase(true)

		if pass1 != pass2 {
			fmt.Println("Phrases are not equal, try again")
		} else {
			return pass1
		}
	}
}

// saveAccountWithPassword performs transactional save of password and account to database
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

// checkSimpleError handles general error
func checkSimpleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

// encryptPassword sets encrypted password and salt for given Password instance
func encryptPassword(password *models.Password, userPassword, encryptionKey string) {
	salt, err := passGenerator.Generate(
		cfg.SaltSettings.Length,
		cfg.SaltSettings.NumDigits,
		cfg.SaltSettings.NumSymbols,
		cfg.SaltSettings.NoUpper,
		cfg.SaltSettings.AllowRepeat)
	checkSimpleError(err, "unable to get salt")

	keyLen := cfg.SecretKeyLength
	key := crypto.DeriveKey(encryptionKey, salt, keyLen)

	encryptedPassword, err := crypto.Encrypt(key, userPassword)
	checkSimpleError(err, "unable to encrypt the password")

	password.Encrypted = encryptedPassword
	password.Salt = salt
}

func PrintServiceRequirements(cfg *config.Config) {
	fmt.Println()
	fmt.Println("For the app to work you need to create the following environment variables:")
	for _, ev := range cfg.GetRequiredEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}

	fmt.Println()

	fmt.Println("You might also want to set the following optional environment variables:")
	for _, ev := range cfg.GetOptionalEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}
	fmt.Println()
}
