package subcommand

import (
	"errors"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/app"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/app/context"
)

func NewEntrypointCommand(entrypointCtx *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:                "entrypoint",
		Short:              "Container entrypoint",
		Long:               "Container entrypoint",
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		Run: func(_ *cobra.Command, _ []string) {
			workflowRunner, err := app.WorkflowRunnerFactory(entrypointCtx)
			if err != nil {
				panic(err)
			}

			defer workflowRunner.Close()

			if !isServiceBuilt() {
				if err := workflowRunner.Execute(commonworkflow.FirstPreStartContainer); err != nil {
					panic(err)
				}
			}

			if err := workflowRunner.Execute(commonworkflow.FirstPreStartService); err != nil {
				panic(err)
			}

			if err := workflowRunner.Execute(commonworkflow.PreStartContainer); err != nil {
				panic(err)
			}

			if err := workflowRunner.Execute(commonworkflow.PreStartService); err != nil {
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
		return syscall.Exec(args[0], args, os.Environ()) // nolint:gosec
	}

	// Shell command
	shellArgs := []string{"/bin/sh", "-c"}
	shellArgs = append(shellArgs, strings.Join(args, " "))

	return syscall.Exec("/bin/sh", shellArgs, os.Environ()) // nolint:gosec
}

func isServiceBuilt() bool {
	markerFile := "/solo/service/data/first_pre_start_container_complete"

	if _, err := os.Stat(markerFile); err != nil {
		return false
	}

	return true
}
