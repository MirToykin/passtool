package cmd

import (
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/atotto/clipboard"
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

		account := requestExistingAccount(&service)
		err := account.LoadPassword(db)
		checkSimpleErrorWithDetails(err, "unable to load account password", printer)

		secret, err := cli.GetSensitiveUserInput("Enter secret: ", printer)
		checkSimpleErrorWithDetails(err, "unable to get secret", printer)

		decoded, err := account.GetDecodedPassword(secret, cfg.SecretKeyLength)
		checkSimpleError(err, "unable to decode password", printer)

		err = clipboard.WriteAll(decoded)
		if err != nil {
			printer.Success("Decoded password: %s", decoded)
		}

		printer.Success("Password copied to clipboard")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
