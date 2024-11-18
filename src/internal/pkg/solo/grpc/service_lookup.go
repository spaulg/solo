package grpc

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/credentials"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type ServiceLookup struct {
	stateDirectory            string
	Hostname                  string
	Port                      uint16
	ClientCertificateFilePath *string `yaml:"client_certificate_file_path,omitempty"`
	ClientKeyFilePath         *string `yaml:"client_key_file_path,omitempty"`
}
type ServiceLookupOptions func(*ServiceLookup)

func NewServiceLookup(hostname string, port uint16, stateDirectory string) *ServiceLookup {
	return &ServiceLookup{
		stateDirectory: stateDirectory,
		Hostname:       hostname,
		Port:           port,
	}
}

func (t *ServiceLookup) ApplyCertificatePack(certificatePack *credentials.CertificatePack) {
	clientCertificate := strings.TrimPrefix(certificatePack.ClientCertificateFilePath, t.stateDirectory+"/")
	t.ClientCertificateFilePath = &clientCertificate

	clientKey := strings.TrimPrefix(certificatePack.ClientCertificateFilePath, t.stateDirectory+"/")
	t.ClientKeyFilePath = &clientKey
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
