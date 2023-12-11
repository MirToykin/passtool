package cmd

import (
	"github.com/MirToykin/passtool/internal/lib/cli"
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
					secret, err := cli.GetSensitiveUserInput("Enter secret: ", deps.printer)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					decoded, err := password.GetDecrypted(secret, deps.config.SecretKeyLength)
					checkSimpleError(err, "unable to decode password", deps.printer)

					err = clipboard.WriteAll(decoded)
					if err != nil {
						deps.printer.Success("Decoded password: %s", decoded)
					}

					deps.printer.Success("Password copied to clipboard")
				},
			)
		},
	}
}

func init() {}
