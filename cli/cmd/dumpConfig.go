package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

// dumpConfigCmd represents the dumpConfig command
var dumpConfigCmd = &cobra.Command{
	Use:   "dump-config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		configYaml, err := yaml.Marshal(config)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to marshall config to yaml: %v", err))
			os.Exit(1)
		}

		fmt.Print(string(configYaml))
	},
}

func init() {
	rootCmd.AddCommand(dumpConfigCmd)
}
