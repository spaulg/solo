package main

import (
	"os"
	"strings"
	"syscall"
)

func main() {
	var err error

	if strings.HasPrefix(os.Args[1], "/") {
		// Full path of executable given
		err = syscall.Exec(os.Args[1], os.Args[1:], nil)
	} else {
		// todo: Requires $PATH env var and needs a shell
		args := []string{"/bin/sh", "-c"}
		args = append(args, strings.Join(os.Args[1:], " "))

		err = syscall.Exec("/bin/sh", args, nil)
	}

	panic(err)
}
