package cmd

import (
	"github.com/MirToykin/passtool/internal/config"

	"github.com/spf13/cobra"
)

// getRequirementsCmd returns the representation of the requirements command
func getRequirementsCmd(config *config.Config, printer Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "requirements",
		Short: "Prints service requirements",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			PrintServiceRequirements(config, printer)
		},
	}
}

func init() {}
