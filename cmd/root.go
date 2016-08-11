package cmd

import (
	"fmt"
	"os"

	"errors"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	inWorkspace = false
	verbose     = false

	//disable color output
	colorOff = false
	// Done informs user about tast completion
	Done = color.New(color.FgHiGreen, color.Bold).PrintfFunc()
	// Warn user about something. PrintfF
	Warn = color.New(color.FgHiYellow, color.Bold).PrintfFunc()
)

type commandError struct {
	cmd       string
	problem   string
	userError bool
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kure",
	Short: "A netkan wrapper that helps you maintain private ckan repo",
	Long: `KURE - Kerbal User REpository
	Tool that makes easy to create and mmaintain you private repository of ckan files.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolVarP(&colorOff, "no-color", "N", false, "Disable colorful output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	//disable color on flag
	if colorOff {
		color.NoColor = true
	}

	viper.SetConfigName("kure") // name of config file (without extension)
	viper.AddConfigPath(".")    // adding current directory
	viper.AutomaticEnv()        // read in environment variables that match

	//defaults
	viper.SetDefault("default_extension", "netkan")
	viper.SetDefault("cachedir", "./cache/download/")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		inWorkspace = true
		if verbose {
			fmt.Println("Using workspace config file:", viper.ConfigFileUsed())
		}
	}
}

func checkWorkspace() error {
	if !inWorkspace {
		return errors.New("You can run this command only from workspace!")
	}
	return nil
}
