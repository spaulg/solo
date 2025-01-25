package main

import (
	"fmt"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/entrypoint"
	"os"
	"path"
	"strings"
	"syscall"
)

func main() {
	workflowRunner, err := entrypoint.WorkflowRunnerFactory()
	if err != nil {
		panic(err)
	}

	defer workflowRunner.Close()

	if !isServiceBuilt() {
		fmt.Println("Executing Build workflow")
		workflowRunner.Execute(commonworkflow.Build)
	}

	fmt.Println("Executing PreStart workflow")
	workflowRunner.Execute(commonworkflow.PreStart)

	fmt.Printf("%+v\n", os.Args)
	err = forkAndExecute(os.Args[1:])
	panic(err)
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
