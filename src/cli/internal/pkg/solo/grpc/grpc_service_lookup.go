package grpc

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type GrpcServiceLookup struct {
	Hostname                  string
	Port                      int
	ClientCertificateFileName string
	ClientKeyFileName         string
}
type GrpcServiceLookupOptions func(*GrpcServiceLookup)

func WithHostname(hostname string) GrpcServiceLookupOptions {
	return func(t *GrpcServiceLookup) {
		t.Hostname = hostname
	}
}

func WithPort(port int) GrpcServiceLookupOptions {
	return func(t *GrpcServiceLookup) {
		t.Port = port
	}
}

func WithClientCertificate(certificateFileName string) GrpcServiceLookupOptions {
	return func(t *GrpcServiceLookup) {
		t.ClientCertificateFileName = certificateFileName
	}
}

func WithClientPrivateKey(keyFileName string) GrpcServiceLookupOptions {
	return func(t *GrpcServiceLookup) {
		t.ClientKeyFileName = keyFileName
	}
}

func NewGrpcServiceLookup(opts ...GrpcServiceLookupOptions) *GrpcServiceLookup {
	t := &GrpcServiceLookup{}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *GrpcServiceLookup) MarshallYaml(filename string) error {
	serviceYaml, err := yaml.Marshal(&t)
	if err != nil {
		return fmt.Errorf("failed to marshall grpc service lookup to yaml: %v", err)
	}

	if err := os.WriteFile(filename, serviceYaml, 0640); err != nil {
		return fmt.Errorf("failed to write grpc service lookup definition file: %v", err)
	}

	return nil
}
