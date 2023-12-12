package cmd

import (
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// getGetCmd returns the representation of the get command
func getGetCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get saved password",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			operation := "get password"
			genericGet(
				operation,
				deps.db,
				deps.printer,
				func(password models.Password) {
					decrypted, err := getDecryptedPasswordWithRetry(password, deps.config.SecretKeyLength, 5, deps.printer)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					err = clipboard.WriteAll(decrypted)
					if err != nil {
						deps.printer.Success("Decoded password: %s", decrypted)
					}

					deps.printer.Success("Password copied to clipboard")
				},
			)
		},
	}
}

func init() {}
