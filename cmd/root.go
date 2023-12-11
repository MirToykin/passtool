package cmd

import (
	"github.com/MirToykin/passtool/cmd/service"
	"github.com/MirToykin/passtool/internal/config"
	"gorm.io/gorm"
	"log"
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

var cmdPrinter Printer
var database *gorm.DB
var appConfig *config.Config

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(dbase *gorm.DB, c *config.Config, printer Printer) {
	database = dbase
	appConfig = c
	cmdPrinter = printer

	sqlDb, err := database.DB()
	if err != nil {
		log.Fatalf("cant get SQL DB: %v", err)
	}

	defer sqlDb.Close()

	err = rootCmd.Execute()
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
	rootCmd.AddCommand(service.AppServiceCmd)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
