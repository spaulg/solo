package cmd

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/project"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"github.com/spaulg/solo/internal/pkg/project_finder"
	"github.com/spf13/cobra"
	"os"
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
		var projectFile *project_file.ProjectFile
		var err error

		if projectFile, err = project_finder.FindProjectFile(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		project := project.New(projectFile)
		project.Destroy()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all solo, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
