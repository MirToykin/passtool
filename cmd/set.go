package cmd

import (
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/atotto/clipboard"
	passGenerator "github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"os"
)

// getSetCmd returns the representation of the set command
func getSetCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set new password for an existing account",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			operation := "set password"
			getPassword, err := getPasswordGetterByGenerateAndLengthFlag(
				cmd, generateFlag, lengthFlag,
				"new password", deps.printer, deps.config)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			genericGet(
				operation,
				deps.db,
				deps.printer,
				func(password models.Password) {
					_, err := getDecryptedPasswordWithRetry(password, deps.config.SecretKeyLength, 5, deps.printer)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					secretKey := getSecretWithConfirmation("secret key for new password", "Secret keys are not equal", deps.printer)
					userPassword, err := getPassword()
					checkSimpleErrorWithDetails(err, operation, deps.printer)
					err = encryptPassword(&password, userPassword, secretKey, deps.config.SecretKeyLength, deps.config.PasswordSettings)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					err = password.Save(deps.db)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					deps.printer.Success("Password updated")

					err = clipboard.WriteAll(userPassword)
					if err == nil {
						deps.printer.Simpleln("Password copied to clipboard")
					}
				},
			)
		},
	}
}

func init() {}

func getGeneratedPassword(length int, config *config.Config, printer Printer) (string, error) {
	if length < config.MinPasswordLength || length > config.MaxPasswordLength {
		printer.Infoln("The password must be at least %d and no more than %d characters long.", config.MinPasswordLength, config.MaxPasswordLength)
		os.Exit(0)
	}

	return passGenerator.Generate(
		length,
		config.PasswordSettings.NumDigits,
		config.PasswordSettings.NumSymbols,
		config.PasswordSettings.NoUpper,
		config.PasswordSettings.AllowRepeat)
}
