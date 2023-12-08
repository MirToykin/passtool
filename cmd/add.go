package cmd

import (
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add your custom password",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		errPrefix := "failed to add"
		var service models.Service
		serviceName := cli.GetUserInput("Enter service name: ", cmdPrinter)
		err := service.FetchOrCreate(db, serviceName)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		var account models.Account
		login, err := requestUniqueLoginForService(&account, service, cmdPrinter)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		account.Service = service
		account.Login = login

		var password models.Password
		userPassword := getSecretWithConfirmation("password", "Passwords are not equal", cmdPrinter)
		secretKey := getSecretWithConfirmation("secret key", "Secret keys are not equal", cmdPrinter)

		err = encryptPassword(&password, userPassword, secretKey)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		err = account.SaveWithPassword(db, &password)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		cmdPrinter.Success("Successfully added password for account with login %q at %q", login, serviceName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
