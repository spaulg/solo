package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewDumpConfigCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "dump-config",
		GroupID: "config",
		Short:   "Dumps the solo config to stdout",
		Long:    "Dumps the solo config to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			configYaml, err := yaml.Marshal(soloCtx.Config)
			if err != nil {
				return err
			}

			fmt.Print(string(configYaml))
			return nil
		},
	}
}
