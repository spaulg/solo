package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/common/version"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewVersionCommand(_ *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Displays the solo version",
		Long:  "Displays the solo version",
		Run: func(_ *cobra.Command, _ []string) {
			info := version.Get()
			fmt.Println(info.String())
		},
	}
}
