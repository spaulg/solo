package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
)

func main() {
	var err error

	fmt.Println("Connect to grpc server")
	port := 12345 // todo: obtain from file stored in bind mount
	// todo: obtain host from orchestrator
	// todo: implement mtls for auth
	conn, err := grpc.NewClient("host.docker.internal:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Creaing new service client")
	client := services.NewProvisionerClient(conn)

	// Signal provisioning run as finished
	fmt.Println("Calling finish")
	_, err = client.NotifyProvisionerComplete(context.Background(), &services.NotifyProvisionerCompleteRequest{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Connect to grpc server")
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
