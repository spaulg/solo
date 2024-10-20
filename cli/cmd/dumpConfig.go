package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// dumpConfigCmd represents the dumpConfig command
var dumpConfigCmd = &cobra.Command{
	Use:   "dump-config",
	Short: "Dumps the solo config to stdout",
	Long:  "Dumps the solo config to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		configYaml, err := yaml.Marshal(config)
		if err != nil {
			return err
		}

		fmt.Print(string(configYaml))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dumpConfigCmd)
}
