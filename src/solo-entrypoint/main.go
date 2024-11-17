package main

import (
	"github.com/spaulg/solo/agent/internal/pkg/entrypoint"
	"os"
	"strings"
	"syscall"
)

func main() {
	provisioner, err := entrypoint.ProvisionerFactory()
	if err != nil {
		panic(err)
	}

	defer provisioner.Close()

	// todo: trigger provisioning actions for particular service

	// Notify provisioner is finished
	provisioner.Finish()

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
