/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/MirToykin/passtool/internal/storage/models"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Prints a list of available services with their accounts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		withAccounts, err := cmd.Flags().GetBool("accounts")
		checkSimpleError(err, "unable to read flags", printer)

		service := models.Service{}
		services, err := service.GetList(db, withAccounts)
		checkSimpleError(err, "unable to get services", printer)

		printServices(services, withAccounts, printer)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("accounts", "a", false, "Print accounts as well")
}
