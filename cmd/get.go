package cmd

import (
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"os"
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

		err := service.List(db).Count(&count).Error
		checkSimpleErrorWithDetails(err, "Unable to check service existence", cmdPrinter)
		if count == 0 {
			cmdPrinter.Infoln("There are no added services yet")
			os.Exit(0)
		}

		err = requestExistingService(&service, cmdPrinter)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)
		printServiceAccounts(service, cmdPrinter)

		account := requestExistingAccount(&service, cmdPrinter)
		err = account.LoadPassword(db)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		secret, err := cli.GetSensitiveUserInput("Enter secret: ", cmdPrinter)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		decoded, err := account.GetDecodedPassword(secret, cfg.SecretKeyLength)
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
