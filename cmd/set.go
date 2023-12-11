package cmd

import (
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set new password for an existing account",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		operation := "set password"
		genericGet(
			operation,
			database,
			cmdPrinter,
			func(password models.Password) {
				secret, err := cli.GetSensitiveUserInput("Enter secret: ", cmdPrinter)
				checkSimpleErrorWithDetails(err, operation, cmdPrinter)

				_, err = password.GetDecrypted(secret, appConfig.SecretKeyLength)
				checkSimpleError(err, "Provided incorrect secret", cmdPrinter)

				userPassword := getSecretWithConfirmation("new password", "Passwords are not equal", cmdPrinter)
				secretKey := getSecretWithConfirmation("secret key", "Secret keys are not equal", cmdPrinter)

				err = encryptPassword(&password, userPassword, secretKey, appConfig.SecretKeyLength, appConfig.PasswordSettings)
				checkSimpleErrorWithDetails(err, operation, cmdPrinter)

				err = password.Save(database)
				checkSimpleErrorWithDetails(err, operation, cmdPrinter)

				cmdPrinter.Success("Password updated")
			},
		)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
