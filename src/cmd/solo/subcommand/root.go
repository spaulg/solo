package subcommand

import (
	"fmt"
	config2 "github.com/spaulg/solo/internal/pkg/solo/config"
	project2 "github.com/spaulg/solo/internal/pkg/solo/project"
	"os"

	"github.com/spf13/cobra"
)

var project *project2.Project
var config *config2.Config
var projectLoadErr, configLoadErr error

var rootCmd = &cobra.Command{
	Use:          "solo",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if projectLoadErr != nil {
			fmt.Println(projectLoadErr)
			os.Exit(1)
		}

		if configLoadErr != nil {
			fmt.Println(configLoadErr)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	config, configLoadErr = config2.NewConfig()
	project, projectLoadErr = project2.FindProject("./")

	if project != nil && configLoadErr == nil {
		configLoadErr = config.AddConfigPath(project.GetDirectory())
	}
}
