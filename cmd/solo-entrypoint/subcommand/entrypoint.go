package subcommand

import (
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/entrypoint/context"
	"github.com/spaulg/solo/internal/pkg/entrypoint/workflow"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"syscall"
)

func NewEntrypointCommand(entrypointCtx *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:                "entrypoint",
		Short:              "Container entrypoint",
		Long:               "Container entrypoint",
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			workflowRunner, err := workflow.WorkflowRunnerFactory(entrypointCtx)
			if err != nil {
				panic(err)
			}

			defer workflowRunner.Close()

			if !isServiceBuilt() {
				workflowRunner.Execute(commonworkflow.FirstPreStart)
			}

			workflowRunner.Execute(commonworkflow.PreStart)

			err = forkAndExecute(os.Args[2:])
			panic(err)
		},
	}
}

func forkAndExecute(args []string) error {
	if []rune(args[0])[0] == '/' {
		// Full path of executable given
		return syscall.Exec(args[0], args, nil)
	} else {
		// Shell command
		shellArgs := []string{"/bin/sh", "-c"}
		shellArgs = append(shellArgs, strings.Join(args, " "))

		return syscall.Exec("/bin/sh", shellArgs, nil)
	}
}

func isServiceBuilt() bool {
	markerFile := path.Join("solo", "service", "build_complete")
	if _, err := os.Stat(markerFile); err != nil {
		return false
	}

	return true
}
