package service

import (
	"github.com/spf13/cobra"
)

// AppServiceCmd represents the service command
var AppServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service requirements",
	Long:  `Prints service usage requirements`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO print service requirements
		_ = cmd.Help()
	},
}

func init() {

}
