package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewSSHCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "ssh",
		GroupID: "tooling",
		Short:   "Drops into a shell on a service, runs commands",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := loadProjectE(soloCtx, []string{}); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ssh called")
		},
	}
}
