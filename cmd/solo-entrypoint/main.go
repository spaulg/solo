package main

import (
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/entrypoint"
	"os"
	"strings"
	"syscall"
)

func main() {
	workflowRunner, err := entrypoint.WorkflowRunnerFactory()
	if err != nil {
		panic(err)
	}

	defer workflowRunner.Close()

	if true { // todo: detect build completed
		workflowRunner.Execute(commonworkflow.Build)
	}

	workflowRunner.Execute(commonworkflow.PreStart)

	err = forkAndExecute(os.Args[1:])
	panic(err)
}

func forkAndExecute(args []string) error {
	if strings.HasPrefix(args[0], "/") {
		// Full path of executable given
		return syscall.Exec(args[0], args, nil)
	} else {
		// todo: Requires $PATH env var and needs a shell
		shellArgs := []string{"/bin/sh", "-c"}
		shellArgs = append(args, strings.Join(args, " "))

		return syscall.Exec("/bin/sh", shellArgs, nil)
	}
}
