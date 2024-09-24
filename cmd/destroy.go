package cmd

import (
	"github.com/spaulg/solo/internal/pkg/project"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		project := project.New(globalConfig, projectFile)
		project.Destroy()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
