package cmd

import (
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"sync"
)

// getAddCmd returns the representation of the add command
func getAddCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add your custom password for a service",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			operation := "add password"
			getPassword, err := getPasswordGetterByGenerateAndLengthFlag(
				cmd, generateFlag, lengthFlag,
				"password", deps.printer, deps.config)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			var service models.Service
			serviceName := cli.GetUserInput("Enter service name: ", deps.printer)
			err = service.FetchOrCreate(deps.db, serviceName)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			var account models.Account
			login, err := requestUniqueLoginForService(&account, service, deps.printer, deps.db)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			account.Service = service
			account.Login = login

			var password models.Password
			userPassword, err := getPassword()
			checkSimpleErrorWithDetails(err, operation, deps.printer)
			secretKey := getSecretWithConfirmation("secret key", "Secret keys are not equal", deps.printer)

			err = encryptPassword(&password, userPassword, secretKey, deps.config.SecretKeyLength, deps.config.PasswordSettings)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			err = account.SaveWithPassword(deps.db, &password)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			wg := sync.WaitGroup{}
			errChan := make(chan error, 2)

			if checkIfBackupNeeded(password.ID, deps.config.BackupIndex) {
				wg.Add(2)
				go createBackup(&wg, deps.config, errChan, deps.printer)
				go clearUnnecessaryBackups(&wg, errChan, deps.config, deps.printer)
			}

			deps.printer.Success("Successfully added password for account with login %q at %q", login, serviceName)

			err = clipboard.WriteAll(userPassword)
			if err == nil {
				deps.printer.Simpleln("Password copied to clipboard")
			}

			wg.Wait()
			close(errChan)

			for err = range errChan {
				if err != nil {
					deps.printer.Warning("failed to handle backup: %v", err)
				}
			}
		},
	}
}

func init() {}
