package cmd

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
)

var projectFile *solo.Project
var globalConfig *solo.Config
var projectLoadErr, globalConfigLoadErr error

// rootCmd represents the base command when called without any solo
var rootCmd = &cobra.Command{
	Use: "solo",
	//Short: "A brief description of your application",
	//Long: `A longer description that spans multiple lines and likely contains
	//examples and usage of using your application. For example:
	//
	//Cobra is a CLI library for Go that empowers applications.
	//This application is a tool to generate the needed files
	//to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if projectLoadErr != nil {
			fmt.Println(projectLoadErr)
			os.Exit(1)
		}

		if globalConfigLoadErr != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(globalConfigLoadErr, &configFileNotFoundError) {
				fmt.Println(globalConfigLoadErr)
				os.Exit(1)
			}
		}
	},
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
	globalConfig, globalConfigLoadErr = solo.NewConfig()
	projectFile, projectLoadErr = solo.FindProjectFile()

	if projectFile != nil {
		globalConfigLoadErr = globalConfig.AddConfigPath(projectFile.Directory)
	}
}
