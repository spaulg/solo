package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/context"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/shells"
)

const shellsFilePath = "/etc/shells"

func NewCatShellsCommand(entrypointCtx *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:   "cat-shells",
		Short: "Cat available shells in json formatted output",
		Long:  "Cat available shells in json formatted output",
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := shells.ListShellsAsJson(shellsFilePath)
			if err != nil {
				return err
			}

			fmt.Println(output)

			return nil
		},
	}
}
