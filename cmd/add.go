package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add your custom password for a service",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		genericAdd(
			"add password",
			database,
			cmdPrinter,
			appConfig,
			func() string {
				return getSecretWithConfirmation("password", "Passwords are not equal", cmdPrinter)
			})
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
