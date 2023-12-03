package cmd

import (
	"errors"
	"fmt"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"log"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add your custom password",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var service models.Service
		serviceName := cli.GetUserInput("Enter service name: ")
		result := db.First(&service, "name", serviceName)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			service.Name = serviceName
			result = db.Create(&service)
		}

		checkSimpleError(result.Error, "unable to create service")

		var account models.Account
		var count int64
		login := cli.GetUserInput("Enter login: ")
		result = db.Model(&account).Where("login = ? AND service_id = ?", login, service.ID).Count(&count)
		checkSimpleError(result.Error, "unable to check account existence")

		if count > 0 {
			log.Fatalf(
				"Account with login %q at %q already exists, to update it use %q command",
				login,
				serviceName,
				updateCmd.Use,
			)
		}

		var password models.Password
		userPassword := cli.GetUserInput("Enter password: ")
		password.Encrypted = userPassword // TODO encryption

		result = db.Create(&password)
		checkSimpleError(result.Error, "unable to save password")

		account.Password = password
		account.Service = service
		account.Login = login

		result = db.Create(&account)
		checkSimpleError(result.Error, "unable to create account")

		fmt.Printf("Successfully added password for account with login %q at %q", login, serviceName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func checkSimpleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
