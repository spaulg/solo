package subcommand

import (
	"fmt"
	config2 "github.com/spaulg/solo/internal/pkg/solo/config"
	project2 "github.com/spaulg/solo/internal/pkg/solo/project"
	"os"

	"github.com/spf13/cobra"
)

type ProjectConfigContext struct {
	Project        *project2.Project
	Config         *config2.Config
	ProjectLoadErr error
	ConfigLoadErr  error
}

func NewRootCommand(projectConfigContext *ProjectConfigContext) *cobra.Command {
	return &cobra.Command{
		Use:          "solo",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if projectConfigContext.ProjectLoadErr != nil {
				fmt.Println(projectConfigContext.ProjectLoadErr)
				os.Exit(1)
			}

			if projectConfigContext.ConfigLoadErr != nil {
				fmt.Println(projectConfigContext.ConfigLoadErr)
				os.Exit(1)
			}
		},
	}
}

func Execute() {
	projectConfigContext := loadConfigAndProject()

	rootCmd := NewRootCommand(projectConfigContext)
	rootCmd.AddCommand(NewDestroySubCommand(projectConfigContext))
	rootCmd.AddCommand(NewDumpComposeConfigCommand(projectConfigContext))
	rootCmd.AddCommand(NewDumpConfigCommand(projectConfigContext))
	rootCmd.AddCommand(NewLogsCommand(projectConfigContext))
	rootCmd.AddCommand(NewRebuildCommand(projectConfigContext))
	rootCmd.AddCommand(NewRestartCommand(projectConfigContext))
	rootCmd.AddCommand(NewSSHCommand(projectConfigContext))
	rootCmd.AddCommand(NewStartCommand(projectConfigContext))
	rootCmd.AddCommand(NewStopCommand(projectConfigContext))
	rootCmd.AddCommand(NewVersionCommand(projectConfigContext))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func loadConfigAndProject() *ProjectConfigContext {
	config, configLoadErr := config2.NewConfig()
	project, projectLoadErr := project2.FindProject("./")

	if project != nil && configLoadErr == nil {
		configLoadErr = config.AddConfigPath(project.GetDirectory())
	}

	return &ProjectConfigContext{
		Config:         config,
		ConfigLoadErr:  configLoadErr,
		Project:        project,
		ProjectLoadErr: projectLoadErr,
	}
}
