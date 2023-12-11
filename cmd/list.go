package cmd

import (
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
func getListCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Prints a list of available services with their accounts",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			operation := "get list"
			withAccounts, err := cmd.Flags().GetBool("accounts")
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			service := models.Service{}
			services, err := service.GetList(deps.db, withAccounts)
			checkSimpleErrorWithDetails(err, operation, deps.printer)

			printServices(services, withAccounts, deps.printer)
		},
	}
}

func init() {}

// printServices prints list of added services and also their accounts if withAccounts=true
func printServices(services []models.Service, withAccounts bool, p Printer) {
	if len(services) == 0 {
		p.Infoln("There are no added services yet")
		return
	}
	p.Header("The following services were added:")
	for i, service := range services {
		p.Infoln("%d. %s", i+1, service.Name)

		if withAccounts {
			for _, account := range service.Accounts {
				p.Simpleln("  - %s", account.Login)
			}
		}
	}
}
