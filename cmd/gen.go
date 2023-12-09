package cmd

import (
	passGenerator "github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate new password for a service",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		genericAdd("generate password",
			db,
			cmdPrinter,
			cfg,
			func() string {
				password, err := passGenerator.Generate(
					cfg.PasswordSettings.Length,
					cfg.PasswordSettings.NumDigits,
					cfg.PasswordSettings.NumSymbols,
					cfg.PasswordSettings.NoUpper,
					cfg.PasswordSettings.AllowRepeat)

				if err != nil {
					checkSimpleErrorWithDetails(err, "failed to generate password", cmdPrinter)
				}

				return password
			})
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}
