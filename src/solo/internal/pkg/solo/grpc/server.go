package grpc

type Server interface {
	Start() error
	Stop()
}
