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
		errPrefix := "failed to list"
		withAccounts, err := cmd.Flags().GetBool("accounts")
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		service := models.Service{}
		services, err := service.GetList(db, withAccounts)
		checkSimpleErrorWithDetails(err, errPrefix, cmdPrinter)

		printServices(services, withAccounts, cmdPrinter)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("accounts", "a", false, "Printer accounts as well")
}
