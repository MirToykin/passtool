package cmd

import (
	"github.com/spf13/cobra"
)

// getAddCmd returns the representation of the add command
func getAddCmd(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add your custom password for a service",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			genericAdd(
				"add password",
				deps,
				func() string {
					return getSecretWithConfirmation("password", "Passwords are not equal", deps.printer)
				})
		},
	}
}

func init() {}
