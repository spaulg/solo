package subcommand

import (
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/app"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

func NewToolCommands(soloCtx *context.CliContext) []*cobra.Command {
	var toolCommands []*cobra.Command

	if soloCtx.Project != nil {
		for toolName, toolConfig := range soloCtx.Project.Tools() {
			localToolName := toolName

			toolCommands = append(toolCommands, &cobra.Command{
				Use:                toolName,
				GroupID:            "tooling",
				Short:              toolConfig.Description,
				Long:               toolConfig.Description,
				DisableFlagParsing: true,
				Annotations: map[string]string{
					RequireConfigLoadSuccessAnnotation:  "true",
					RequireProjectLoadSuccessAnnotation: "true",
				},
				RunE: func(_ *cobra.Command, args []string) error {
					projectTooling, err := app.ProjectToolingFactory(soloCtx)
					if err != nil {
						return err
					}

					return projectTooling.ExecuteTool(localToolName, args)
				},
			})
		}
	}

	return toolCommands
}
