package grpc

func StartServer() (int, error) {
	// Start the GRPC server and return the port number the server is listening on
	grpcServicePortChannel := make(chan int)
	grpcServiceErrorChannel := make(chan error)

	go func() {
		grpcServer := NewGrpcServer()

		// Start listener and report listening port
		port, err := grpcServer.CreateListener()
		if err != nil {
			grpcServiceErrorChannel <- err

			close(grpcServicePortChannel)
			close(grpcServiceErrorChannel)

			return
		}

		grpcServicePortChannel <- port

		// Start listening
		grpcServer.Listen()

		close(grpcServicePortChannel)
		close(grpcServiceErrorChannel)
	}()

	port, ok := <-grpcServicePortChannel
	if !ok {
		return 0, <-grpcServiceErrorChannel
	}

	return port, nil
}
