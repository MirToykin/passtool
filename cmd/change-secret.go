package cmd

import (
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// getDelCmd returns the representation of the change-secret command
func getChangeSecretCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "change-secret",
		Short: "Set new secret key for a password",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			operation := "change secret"
			getHandler := func() func(account models.Account) {
				return func(account models.Account) {
					password := account.Password
					decrypted, err := getDecryptedPasswordWithRetry(password, deps.config.SecretKeyLength, 5, deps.printer)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					secretKey := getSecretWithConfirmation("new secret key", "Secret keys are not equal", deps.printer)
					err = encryptPassword(&password, decrypted, secretKey, deps.config.SecretKeyLength, deps.config.PasswordSettings)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					err = password.Save(deps.db)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					deps.printer.Success("Secret key updated")
				}
			}
			genericGet(
				operation,
				deps.db,
				deps.printer,
				getHandler(),
			)
		},
	}
}

func init() {}
