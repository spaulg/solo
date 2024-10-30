package grpc

func StartServer(certificateFilePath string, certificateKeyPath string, caCertificateFilePath string) (int, error) {
	grpcServicePortChannel := make(chan int)
	grpcServiceErrorChannel := make(chan error)

	go func() {
		grpcServer := NewGrpcServer(certificateFilePath, certificateKeyPath, caCertificateFilePath)

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
		_ = grpcServer.Listen()

		close(grpcServicePortChannel)
		close(grpcServiceErrorChannel)
	}()

	port, ok := <-grpcServicePortChannel
	if !ok {
		return 0, <-grpcServiceErrorChannel
	}

	return port, nil
}
