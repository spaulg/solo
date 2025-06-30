package subcommand

import (
	"errors"
	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/context"
	"github.com/spf13/cobra"
	"os"
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
			workflowRunner, err := entrypoint.WorkflowRunnerFactory(entrypointCtx)
			if err != nil {
				panic(err)
			}

			defer workflowRunner.Close()

			if !isServiceBuilt() {
				if err := workflowRunner.Execute(commonworkflow.FirstPreStart); err != nil {
					panic(err)
				}
			}

			if err := workflowRunner.Execute(commonworkflow.PreStart); err != nil {
				panic(err)
			}

			err = forkAndExecute(os.Args[2:])
			panic(err)
		},
	}
}

func forkAndExecute(args []string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return errors.New("no command specified")
	}

	if []rune(args[0])[0] == '/' {
		// Full path of executable given
		return syscall.Exec(args[0], args, os.Environ())
	} else {
		// Shell command
		shellArgs := []string{"/bin/sh", "-c"}
		shellArgs = append(shellArgs, strings.Join(args, " "))

		return syscall.Exec("/bin/sh", shellArgs, os.Environ())
	}
}

func isServiceBuilt() bool {
	markerFile := "/solo/service/data/first_pre_start_complete"

	if _, err := os.Stat(markerFile); err != nil {
		return false
	}

	return true
}
