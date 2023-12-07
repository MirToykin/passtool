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
	err := service.FetchByName(db, serviceName, false)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		service.Name = serviceName
		err = db.Create(&service).Error
	}

	checkSimpleError(err, "unable to create service")
}

// fetchService tries to fetch service
func fetchService(service *models.Service, serviceName string) bool {
	err := service.FetchByName(db, serviceName, true)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}

	checkSimpleError(err, "unable to get service")
	return true
}

// getAccountsCount returns number of accounts with a given login for a given service
func getAccountsCount(account *models.Account, login string, serviceID uint) int64 {
	var count int64
	result := db.Model(&account).Where("login = ? AND service_id = ?", login, serviceID).Count(&count)
	checkSimpleError(result.Error, "unable to check account existence")
	return count
}

// requestUniqueLogin request login from user. If login already exists for the given service - retries.
func requestUniqueLogin(account *models.Account, serviceID uint, serviceName string) string {
	for {
		login := cli.GetUserInput("Enter login: ", printer)
		count := getAccountsCount(account, login, serviceID)

		if count > 0 {
			log.Printf(
				"Account with login %q at %q already exists, to update it use %q command. Use another login.",
				login,
				serviceName,
				updateCmd.Use,
			)
		} else {
			return login
		}
	}
}

// requestExistingLogin request login from user. If login doesn't exist for the given service - retries.
// If succeeds - loads account by login
func requestExistingAccount(account *models.Account, service models.Service) {
	for {
		login := cli.GetUserInput("Enter login: ", printer)
		err := account.FetchByLoginAndService(db, login, service.ID)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			printer.Warning(
				"Account with login %q doesn't exist at service %q. Use another login.",
				login,
				service.Name,
			)
		} else {
			checkSimpleError(err, "unable to account")
			return
		}

	}
}

// getPassPhrase returns pass phrase given by user (handle possible errors)
func getPassPhrase(confirm bool) string {
	postfix := ""
	if confirm {
		postfix = " again"
	}

	secretKey, err := cli.GetSensitiveUserInput(fmt.Sprintf("Enter secret %s: ", postfix), printer)
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
		printer.ErrorWithExit("%s: %v", msg, err)
	}
}

// encryptPassword sets encrypted password and salt for given Password instance
func encryptPassword(password *models.Password, userPassword, secret string) {
	salt, err := passGenerator.Generate(
		cfg.SaltSettings.Length,
		cfg.SaltSettings.NumDigits,
		cfg.SaltSettings.NumSymbols,
		cfg.SaltSettings.NoUpper,
		cfg.SaltSettings.AllowRepeat)
	checkSimpleError(err, "unable to get salt")

	keyLen := cfg.SecretKeyLength
	key := crypto.DeriveKey(secret, salt, keyLen)

	encryptedPassword, err := crypto.Encrypt(key, userPassword)
	checkSimpleError(err, "unable to encrypt the password")

	password.Encrypted = encryptedPassword
	password.Salt = salt
}

// printServiceAccounts prints accounts of the given service into console
func printServiceAccounts(service models.Service) {
	printer.Header("Service %q has accounts with the following logins:", service.Name)
	for i, account := range service.Accounts {
		fmt.Println(i+1, account.Login)
	}
	fmt.Println()
}

// requestExistingService queries for an existing service name until it gets one.
// If succeeds - loads service by name
func requestExistingService(service *models.Service) {
	for {
		serviceName := cli.GetUserInput("Enter service name: ", printer)

		ok := fetchService(service, serviceName)
		if !ok {
			printer.Warning("Service with name %q not found, try again", serviceName)
		} else {
			return
		}
	}
}

// PrintServiceRequirements prints the information for service to be able to work
func PrintServiceRequirements(cfg *config.Config, printer Print) {
	fmt.Println()
	printer.Info("For the app to work you need to add the following environment variables:")
	for _, ev := range cfg.GetRequiredEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}

	fmt.Println()

	printer.Info("You might also want to add the following optional environment variables:")
	for _, ev := range cfg.GetOptionalEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}
	fmt.Println()
}
