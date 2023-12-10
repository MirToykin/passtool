package cmd

import (
	"errors"
	"fmt"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"os"
	"sort"
	"strconv"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get saved password",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var service models.Service
		var count int64
		errPrefix := "failed to add"

		err := service.List(database).Count(&count).Error
		checkSimpleErrorWithDetails(err, "Unable to check service existence", cmdPrinter)
		if count == 0 {
			cmdPrinter.Infoln("There are no added services yet")
			os.Exit(0)
		}

		err = requestExistingService(database, &service, cmdPrinter)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)
		printServiceAccounts(service, cmdPrinter)

		account := requestExistingAccount(&service, cmdPrinter)
		err = account.LoadPassword(database)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		secret, err := cli.GetSensitiveUserInput("Enter secret: ", cmdPrinter)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		decoded, err := account.GetDecodedPassword(secret, appConfig.SecretKeyLength)
		checkSimpleError(err, "unable to decode password", cmdPrinter)

		err = clipboard.WriteAll(decoded)
		if err != nil {
			cmdPrinter.Success("Decoded password: %s", decoded)
		}

		cmdPrinter.Success("Password copied to clipboard")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

// requestExistingService queries for an existing service name until it gets one.
// If succeeds - loads service with its accounts
func requestExistingService(db *gorm.DB, service *models.Service, p Printer) error {
	for {
		serviceName := cli.GetUserInput("Enter service name: ", p)

		ok, err := fetchServiceWithAccounts(db, service, serviceName)
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

// fetchServiceWithAccounts tries to fetch service
func fetchServiceWithAccounts(db *gorm.DB, service *models.Service, serviceName string) (bool, error) {
	err := service.FetchByName(db, serviceName, true)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to fetch service: %w", err)
	}

	return true, nil
}

// requestExistingAccount request login from user. If login doesn't exist for the given service - retries.
// If succeeds - loads account by login
func requestExistingAccount(service *models.Service, printer Printer) *models.Account {
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
