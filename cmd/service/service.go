package service

import (
	"github.com/spf13/cobra"
)

// ServiceCmd represents the service command
var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service requirements",
	Long:  `Prints service usage requirements`,
	Run: func(cmd *cobra.Command, args []string) {

		_ = cmd.Help()
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
