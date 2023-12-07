package cmd

import (
	"fmt"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get saved password",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var service models.Service
		requestExistingService(&service)
		printServiceAccounts(service)

		var account models.Account
		requestExistingAccount(&account, service)

		secret, err := cli.GetSensitiveUserInput("Enter secret: ", printer)
		checkSimpleError(err, "unable to get secret")
		decoded, err := account.GetDecodedPassword(secret, cfg.SecretKeyLength)
		checkSimpleError(err, "unable to decode password")

		fmt.Println(decoded)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
