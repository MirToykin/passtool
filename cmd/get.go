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
	Short: "Get saved password",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		operation := "get password"
		genericGet(
			operation,
			database,
			cmdPrinter,
			func(password models.Password) {
				secret, err := cli.GetSensitiveUserInput("Enter secret: ", cmdPrinter)
				checkSimpleErrorWithDetails(err, operation, cmdPrinter)

				decoded, err := password.GetDecrypted(secret, appConfig.SecretKeyLength)
				checkSimpleError(err, "unable to decode password", cmdPrinter)

				err = clipboard.WriteAll(decoded)
				if err != nil {
					cmdPrinter.Success("Decoded password: %s", decoded)
				}

				cmdPrinter.Success("Password copied to clipboard")
			},
		)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
