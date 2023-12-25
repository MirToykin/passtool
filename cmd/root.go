package cmd

import (
	"github.com/MirToykin/passtool/internal/config"
	out "github.com/MirToykin/passtool/internal/output"
	"github.com/MirToykin/passtool/internal/storage"
	"gorm.io/gorm"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "passtool",
	Short: "Tool for password managing",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

const (
	generateFlag = "generate"
	lengthFlag   = "length"
)

type GenSettings interface {
	GetLength() int
	GetNumDigits() int
	GetNumSymbols() int
	GetNoUpper() bool
	GetAllowRepeat() bool
}

type Printer interface {
	Simple(msg string, a ...interface{})
	Simpleln(msg string, a ...interface{})
	Info(msg string, a ...interface{})
	Infoln(msg string, a ...interface{})
	Success(msg string, a ...interface{})
	Header(msg string, a ...interface{})
	Warning(msg string, a ...interface{})
	Error(msg string, a ...interface{})
	ErrorWithExit(msg string, a ...interface{})
}

type AppDependencies struct {
	db      *gorm.DB
	config  *config.Config
	printer Printer
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.passtool.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	printer := out.New()
	cfg := config.Load()
	if !cfg.IsValid() {
		PrintServiceRequirements(cfg, printer)
		os.Exit(0)
	}

	db, err := storage.New(cfg.StoragePath)
	if err != nil {
		printer.ErrorWithExit("unable to initialize DB: %v", err)
	}

	dependencies := AppDependencies{
		db:      db,
		config:  cfg,
		printer: printer,
	}

	// ============== Register commands ==================

	// requirements
	rootCmd.AddCommand(getRequirementsCmd(dependencies.config, dependencies.printer))

	// add
	addCmd := getAddCmd(dependencies)
	setGenerationFlags(addCmd, dependencies.config.PasswordSettings.Length)
	rootCmd.AddCommand(addCmd)

	// get
	rootCmd.AddCommand(getGetCmd(dependencies))

	// del
	rootCmd.AddCommand(getDelCmd(dependencies))

	// set
	setCmd := getSetCmd(dependencies)
	setGenerationFlags(setCmd, dependencies.config.PasswordSettings.Length)
	rootCmd.AddCommand(setCmd)

	// change-secret
	rootCmd.AddCommand(getChangeSecretCmd(dependencies))

	// list
	listCmd := getListCmd(dependencies)
	listCmd.Flags().BoolP("accounts", "a", false, "Print accounts as well")
	rootCmd.AddCommand(listCmd)
}

// setGenerationFlags sets flags related to password generation to the given command
func setGenerationFlags(cmd *cobra.Command, defaultLength int) {
	cmd.Flags().BoolP(generateFlag, "g", false, "Generate secure password")
	cmd.Flags().Int(lengthFlag, defaultLength, "Length of generated password")
}
