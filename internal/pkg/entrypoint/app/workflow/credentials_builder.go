package workflow

import "google.golang.org/grpc/credentials"

type Builder interface {
	Build() (credentials.TransportCredentials, error)
}
