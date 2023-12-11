package cmd

import (
	passGenerator "github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

// getGenCommand returns the representation of the gen command
func getGenCommand(deps AppDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "gen",
		Short: "Generate new password for a service",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			length := deps.config.PasswordSettings.Length
			if len(args) > 0 {
				l, err := strconv.Atoi(args[0])
				if err != nil {
					deps.printer.ErrorWithExit("Invalid argument provided for password length")
				}

				if l < deps.config.MinPasswordLength || l > deps.config.MaxPasswordLength {
					deps.printer.Infoln("The password must be at least %d and no more than %d characters long.", deps.config.MinPasswordLength, deps.config.MaxPasswordLength)
					os.Exit(0)
				}

				length = l
			}
			genericAdd("generate password",
				deps,
				func() string {
					password, err := passGenerator.Generate(
						length,
						deps.config.PasswordSettings.NumDigits,
						deps.config.PasswordSettings.NumSymbols,
						deps.config.PasswordSettings.NoUpper,
						deps.config.PasswordSettings.AllowRepeat)

					if err != nil {
						checkSimpleErrorWithDetails(err, "failed to generate password", deps.printer)
					}

					return password
				})
		},
	}
}

func init() {}
