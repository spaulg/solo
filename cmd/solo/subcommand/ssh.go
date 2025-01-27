package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewSSHCommand(_ *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "ssh",
		GroupID: "tooling",
		Short:   "Drops into a shell on a service, runs commands",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ssh called")
		},
	}
}
