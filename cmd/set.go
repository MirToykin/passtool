package cmd

import (
	"fmt"
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
	passGenerator "github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"os"
)

const (
	generateFlag  = "generate"
	lengthFlag    = "length"
	lengthDefault = 12
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set new password for an existing account",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		operation := "set password"
		needGenerate, err := cmd.Flags().GetBool(generateFlag)
		checkSimpleErrorWithDetails(err, operation, cmdPrinter)

		var userPassword string
		if needGenerate {
			length, err := cmd.Flags().GetInt(lengthFlag)
			checkSimpleErrorWithDetails(err, operation, cmdPrinter)

			userPassword, err = getGeneratedPassword(length, appConfig)
			checkSimpleErrorWithDetails(err, "failed to generate password", cmdPrinter)
		} else {
			userPassword = getSecretWithConfirmation("new password", "Passwords are not equal", cmdPrinter)
		}

		genericGet(
			operation,
			database,
			cmdPrinter,
			func(password models.Password) {
				secret, err := cli.GetSensitiveUserInput("Enter secret: ", cmdPrinter)
				checkSimpleErrorWithDetails(err, operation, cmdPrinter)

				_, err = password.GetDecrypted(secret, appConfig.SecretKeyLength)
				checkSimpleError(err, "Provided incorrect secret", cmdPrinter)

				secretKey := getSecretWithConfirmation("secret key for new password", "Secret keys are not equal", cmdPrinter)

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

	setCmd.Flags().BoolP(generateFlag, "g", false, "Generate secure password")
	setCmd.Flags().Int(lengthFlag, lengthDefault, fmt.Sprintf("Length of generated password, by default %d", lengthDefault))
}

func getGeneratedPassword(length int, config *config.Config) (string, error) {
	if length < config.MinPasswordLength || length > config.MaxPasswordLength {
		cmdPrinter.Infoln("The password must be at least %d and no more than %d characters long.", config.MinPasswordLength, config.MaxPasswordLength)
		os.Exit(0)
	}

	return passGenerator.Generate(
		length,
		config.PasswordSettings.NumDigits,
		config.PasswordSettings.NumSymbols,
		config.PasswordSettings.NoUpper,
		config.PasswordSettings.AllowRepeat)
}
