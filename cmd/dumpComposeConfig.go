package cmd

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/config"
	"github.com/spaulg/solo/internal/pkg/project"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"github.com/spaulg/solo/internal/pkg/project_finder"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
)

// composeConfigCmd represents the composeConfig command
var composeConfigCmd = &cobra.Command{
	Use:   "dump-compose-config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var projectFile *project_file.ProjectFile
		var err error

		// Find project file
		if projectFile, err = project_finder.FindProjectFile(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Read project configuration
		projectConfig, err := config.ReadConfig(projectFile)
		if err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		// Action project
		project := project.New(projectConfig, projectFile)
		project.DumpComposeConfig()
	},
}

func init() {
	rootCmd.AddCommand(composeConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// composeConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// composeConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
