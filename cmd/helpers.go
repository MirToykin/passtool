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

// fetchServiceWithAccounts tries to fetch service
func fetchServiceWithAccounts(service *models.Service, serviceName string) (bool, error) {
	err := service.FetchByName(db, serviceName, true)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to fetch service: %w", err)
	}

	return true, nil
}

// printServices prints list of added services and also their accounts if withAccounts=true
func printServices(services []models.Service, withAccounts bool, p Printer) {
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

// requestUniqueLoginForService request login from user. If login already exists for the given service - retries.
func requestUniqueLoginForService(account *models.Account, service models.Service, printer Printer) (string, error) {
	for {
		login := cli.GetUserInput("Enter login: ", printer)

		var count int64
		err := account.FindByLoginAndServiceID(db, login, service.ID).Count(&count).Error
		if err != nil {
			return "", fmt.Errorf("faild to request unique login: %w", err)
		}

		if count > 0 {
			log.Printf(
				"Account with login %q at %q already exists, to update it use %q command. Use another login.",
				login,
				service.Name,
				updateCmd.Use,
			)
		} else {
			return login, nil
		}
	}
}

// requestExistingAccount request login from user. If login doesn't exist for the given service - retries.
// If succeeds - loads account by login
func requestExistingAccount(service *models.Service, p Printer) *models.Account {
	for {
		identifier := cli.GetUserInput("Enter login or serial number: ", p)
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

		p.Warning(
			"Account with login %q doesn't exist at service %q. Use another login or correct serial number.",
			login,
			service.Name,
		)
	}
}

// getSecret returns secret given by user (handle possible errors)
func getSecret(secretName string, confirm bool, printer Printer) string {
	postfix := ""
	if confirm {
		postfix = " again"
	}

	secret, err := cli.GetSensitiveUserInput(fmt.Sprintf("Enter %s%s: ", secretName, postfix), printer)
	checkSimpleErrorWithDetails(err, fmt.Sprintf("unable to get %s", secretName), printer)
	return secret
}

// getSecretWithConfirmation handles getting pass phrase with confirmation
func getSecretWithConfirmation(secretName string, retryMsg string, printer Printer) string {
	for {
		pass1 := getSecret(secretName, false, printer)
		pass2 := getSecret(secretName, true, printer)

		if pass1 != pass2 {
			fmt.Println(retryMsg)
		} else {
			return pass1
		}
	}
}

// checkSimpleErrorWithDetails handles general error, and prints message with error details to the console
func checkSimpleErrorWithDetails(err error, msg string, p Printer) {
	if err != nil {
		p.ErrorWithExit("%s: %v", msg, err)
	}
}

// checkSimpleError handles general error and prints message to the console.
func checkSimpleError(err error, msg string, p Printer) {
	if err != nil {
		p.ErrorWithExit(msg)
	}
}

// encryptPassword sets encrypted password and salt for given Password instance
func encryptPassword(password *models.Password, userPassword, secret string) error {
	salt, err := passGenerator.Generate(
		cfg.SaltSettings.Length,
		cfg.SaltSettings.NumDigits,
		cfg.SaltSettings.NumSymbols,
		cfg.SaltSettings.NoUpper,
		cfg.SaltSettings.AllowRepeat)

	if err != nil {
		return fmt.Errorf("unable to get salt: %w", err)
	}

	keyLen := cfg.SecretKeyLength
	key := crypto.DeriveKey(secret, salt, keyLen)

	encryptedPassword, err := crypto.Encrypt(key, userPassword)
	if err != nil {
		return fmt.Errorf("unable to encrypt the password: %w", err)
	}

	password.Encrypted = encryptedPassword
	password.Salt = salt
	return nil
}

// printServiceAccounts prints accounts of the given service into console
func printServiceAccounts(service models.Service, p Printer) {
	p.Header("Service %q has accounts with the following logins:", service.Name)
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
func requestExistingService(service *models.Service, p Printer) error {
	for {
		serviceName := cli.GetUserInput("Enter service name: ", p)

		ok, err := fetchServiceWithAccounts(service, serviceName)
		if err != nil {
			return fmt.Errorf("failed to request existing service: %w", err)
		}

		if !ok {
			p.Warning("Service with name %q not found, try again", serviceName)
		} else {
			return nil
		}
	}
}

// PrintServiceRequirements prints the information for service to be able to work
func PrintServiceRequirements(cfg *config.Config, p Printer) {
	fmt.Println()
	p.Info("For the app to work you need to add the following environment variables:")
	for _, ev := range cfg.GetRequiredEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}

	fmt.Println()

	p.Info("You might also want to add the following optional environment variables:")
	for _, ev := range cfg.GetOptionalEnvVars() {
		fmt.Println(fmt.Sprintf("  %q - %s", ev.Name, ev.Description))
	}
	fmt.Println()
}
