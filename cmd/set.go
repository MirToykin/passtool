package cmd

import (
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/lib/cli"
	"github.com/MirToykin/passtool/internal/storage/models"
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
			needGenerate, err := cmd.Flags().GetBool(generateFlag)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			var userPassword string
			if needGenerate {
				length, err := cmd.Flags().GetInt(lengthFlag)
				checkSimpleErrorWithDetails(err, operation, deps.printer)

				userPassword, err = getGeneratedPassword(length, deps.config, deps.printer)
				checkSimpleErrorWithDetails(err, "failed to generate password", deps.printer)
			} else {
				userPassword = getSecretWithConfirmation("new password", "Passwords are not equal", deps.printer)
			}

			genericGet(
				operation,
				deps.db,
				deps.printer,
				func(password models.Password) {
					secret, err := cli.GetSensitiveUserInput("Enter secret: ", deps.printer)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					_, err = password.GetDecrypted(secret, deps.config.SecretKeyLength)
					checkSimpleError(err, "Provided incorrect secret", deps.printer)

					secretKey := getSecretWithConfirmation("secret key for new password", "Secret keys are not equal", deps.printer)

					err = encryptPassword(&password, userPassword, secretKey, deps.config.SecretKeyLength, deps.config.PasswordSettings)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					err = password.Save(deps.db)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					deps.printer.Success("Password updated")
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
