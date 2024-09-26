package cmd

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
)

// rebuildCmd represents the rebuild command
var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		project := solo.NewProject(globalConfig, projectFile)
		project.Destroy()
		project.Start()
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}
