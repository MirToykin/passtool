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
	"sort"
	"strconv"
)

// fetchOrCreateService fetches existing or creates new Service instance
func fetchOrCreateService(service *models.Service, serviceName string) {
	err := service.FetchByName(db, serviceName, false)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		service.Name = serviceName
		err = db.Create(&service).Error
	}

	checkSimpleErrorWithDetails(err, "unable to create service", printer)
}

// fetchServiceWithAccounts tries to fetch service
func fetchServiceWithAccounts(service *models.Service, serviceName string) bool {
	err := service.FetchByName(db, serviceName, true)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}

	checkSimpleErrorWithDetails(err, "unable to get service", printer)
	return true
}

// printServices prints list of added services and also their accounts if withAccounts=true
func printServices(services []models.Service, withAccounts bool, p Print) {
	if len(services) == 0 {
		p.Infoln("There are no added services yet")
		return
	}
	p.Header("The following services were added:")
	for i, service := range services {
		p.Infoln("%d. %s", i+1, service.Name)

		if withAccounts {
			for _, account := range service.Accounts {
				p.Simpleln("  - %s", account.Login)
			}
		}
	}
}

// getAccountsCount returns number of accounts with a given login for a given service
func getAccountsCount(account *models.Account, login string, serviceID uint) int64 {
	var count int64
	result := db.Model(&account).Where("login = ? AND service_id = ?", login, serviceID).Count(&count)
	checkSimpleErrorWithDetails(result.Error, "unable to check account existence", printer)
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

// requestExistingAccount request login from user. If login doesn't exist for the given service - retries.
// If succeeds - loads account by login
func requestExistingAccount(service *models.Service) *models.Account {
	for {
		identifier := cli.GetUserInput("Enter login or serial number: ", printer)
		var login string

		num, err := strconv.Atoi(identifier)
		if err != nil {
			login = identifier
		} else {
			account, found := service.GetAccountsMap()[num]
			if found {
				return &account
			} else {
				login = identifier
			}
		}

		for _, acc := range service.Accounts {
			if acc.Login == login {
				return &acc
			}
		}

		printer.Warning(
			"Account with login %q doesn't exist at service %q. Use another login or correct serial number.",
			login,
			service.Name,
		)
	}
}

// getPassPhrase returns pass phrase given by user (handle possible errors)
func getPassPhrase(confirm bool) string {
	postfix := ""
	if confirm {
		postfix = " again"
	}

	secretKey, err := cli.GetSensitiveUserInput(fmt.Sprintf("Enter secret %s: ", postfix), printer)
	checkSimpleErrorWithDetails(err, "unable to get passphrase", printer)
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
	checkSimpleErrorWithDetails(err, "unable to create password", printer)
}

// checkSimpleErrorWithDetails handles general error, and prints message with error details to the console
func checkSimpleErrorWithDetails(err error, msg string, p Print) {
	if err != nil {
		p.ErrorWithExit("%s: %v", msg, err)
	}
}

// checkSimpleError handles general error and prints message to the console.
func checkSimpleError(err error, msg string, p Print) {
	if err != nil {
		p.ErrorWithExit(msg)
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
	checkSimpleErrorWithDetails(err, "unable to get salt", printer)

	keyLen := cfg.SecretKeyLength
	key := crypto.DeriveKey(secret, salt, keyLen)

	encryptedPassword, err := crypto.Encrypt(key, userPassword)
	checkSimpleErrorWithDetails(err, "unable to encrypt the password", printer)

	password.Encrypted = encryptedPassword
	password.Salt = salt
}

// printServiceAccounts prints accounts of the given service into console
func printServiceAccounts(service models.Service) {
	printer.Header("Service %q has accounts with the following logins:", service.Name)
	accMap := service.GetAccountsMap()
	keys := make([]int, 0, len(accMap))

	for key := range accMap {
		keys = append(keys, key)
	}

	sort.Ints(keys)
	for _, key := range keys {
		fmt.Println(key, accMap[key].Login)
	}

	fmt.Println()
}

// requestExistingService queries for an existing service name until it gets one.
// If succeeds - loads service with its accounts
func requestExistingService(service *models.Service) {
	for {
		serviceName := cli.GetUserInput("Enter service name: ", printer)

		ok := fetchServiceWithAccounts(service, serviceName)
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
