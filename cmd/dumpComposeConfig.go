package cmd

import (
	"github.com/spaulg/solo/internal/pkg/project"
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
		project := project.LoadProject(globalConfig, projectFile)
		project.DumpComposeConfig()
	},
}

func init() {
	rootCmd.AddCommand(composeConfigCmd)
}
