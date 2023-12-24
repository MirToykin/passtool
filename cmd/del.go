package cmd

import (
	"fmt"
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// getDelCmd returns the representation of the del command
func getDelCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "del",
		Short: "Delete saved password",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			operation := "delete password"
			getHandler := func() func(account models.Account) {
				return func(account models.Account) {
					_, err := getDecryptedPasswordWithRetry(account.Password, deps.config.SecretKeyLength, 5, deps.printer)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					err = account.DeleteWithPassword(deps.db)
					checkSimpleErrorWithDetails(err, operation, deps.printer)
					deps.printer.Success("Account and password deleted")

					service := account.Service
					accountsCount, err := service.AccountsCount(deps.db)
					checkSimpleErrorWithDetails(err, operation, deps.printer)

					if accountsCount == 0 {
						err = service.Delete(deps.db)
						checkSimpleErrorWithDetails(err, operation, deps.printer)
						deps.printer.Success(fmt.Sprintf("The service %q has been deleted because it has no accounts", service.Name))
					}
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
