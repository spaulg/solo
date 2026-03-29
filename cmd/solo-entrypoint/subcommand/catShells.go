package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/infra/shells"
)

const shellsFilePath = "/etc/shells"

func NewCatShellsCommand(_ *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:   "cat-shells",
		Short: "Cat available shells in json formatted output",
		Long:  "Cat available shells in json formatted output",
		RunE: func(_ *cobra.Command, _ []string) error {
			output, err := shells.ListShellsAsJSON(shellsFilePath)
			if err != nil {
				return err
			}

			fmt.Println(output)

			return nil
		},
	}
}
