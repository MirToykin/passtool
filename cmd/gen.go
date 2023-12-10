package cmd

import (
	passGenerator "github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate new password for a service",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		length := cfg.PasswordSettings.Length
		if len(args) > 0 {
			l, err := strconv.Atoi(args[0])
			if err != nil {
				cmdPrinter.ErrorWithExit("Invalid argument provided for password length")
			}

			if l < cfg.MinGeneratedPasswordLength {
				cmdPrinter.Infoln("Password length should be at least %d symbols", cfg.MinGeneratedPasswordLength)
				os.Exit(0)
			}

			length = l
		}
		genericAdd("generate password",
			db,
			cmdPrinter,
			cfg,
			func() string {
				password, err := passGenerator.Generate(
					length,
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
