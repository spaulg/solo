package grpc

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type ServiceLookup struct {
	Hostname                  string
	Port                      uint16
	ClientCertificateFileName string
	ClientKeyFileName         string
}
type ServiceLookupOptions func(*ServiceLookup)

func WithHostname(hostname string) ServiceLookupOptions {
	return func(t *ServiceLookup) {
		t.Hostname = hostname
	}
}

func WithPort(port uint16) ServiceLookupOptions {
	return func(t *ServiceLookup) {
		t.Port = port
	}
}

func WithClientCertificate(certificateFileName string) ServiceLookupOptions {
	return func(t *ServiceLookup) {
		t.ClientCertificateFileName = certificateFileName
	}
}

func WithClientPrivateKey(keyFileName string) ServiceLookupOptions {
	return func(t *ServiceLookup) {
		t.ClientKeyFileName = keyFileName
	}
}

func NewServiceLookup(opts ...ServiceLookupOptions) *ServiceLookup {
	t := &ServiceLookup{}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *ServiceLookup) MarshallYaml(filename string) error {
	serviceYaml, err := yaml.Marshal(&t)
	if err != nil {
		return fmt.Errorf("failed to marshall grpc service lookup to yaml: %v", err)
	}

	if err := os.WriteFile(filename, serviceYaml, 0640); err != nil {
		return fmt.Errorf("failed to write grpc service lookup definition file: %v", err)
	}

	return nil
}
