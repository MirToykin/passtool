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
		var service models.Service
		serviceName := cli.GetUserInput("Enter service name: ", printer)
		fetchOrCreateService(&service, serviceName)

		var account models.Account
		login := requestUniqueLogin(&account, service.ID, serviceName)

		account.Service = service
		account.Login = login

		var password models.Password
		userPassword := getSecretWithConfirmation("password", "Passwords are not equal")
		secretKey := getSecretWithConfirmation("secret key", "Secret keys are not equal")

		encryptPassword(&password, userPassword, secretKey)
		saveAccountWithPassword(&account, &password)

		printer.Success("Successfully added password for account with login %q at %q", login, serviceName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
