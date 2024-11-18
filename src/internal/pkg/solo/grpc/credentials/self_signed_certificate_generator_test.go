package credentials

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"testing"
)

func TestEmptyHostname(t *testing.T) {
	_, err := NewCertificateGenerator("", t.TempDir())
	if err == nil {
		t.FailNow()
	}

	if errors.Is(err, InvalidHostname) {
		return
	}

	t.FailNow()
}

func TestEmptyCertificatePath(t *testing.T) {
	_, err := NewCertificateGenerator("example.com", "")
	if err == nil {
		t.FailNow()
	}

	if errors.Is(err, InvalidCertificateBasePath) {
		return
	}

	t.FailNow()
}

func TestCertificateGeneration(t *testing.T) {
	certificateGenerator, err := NewCertificateGenerator("example.com", t.TempDir())
	if err != nil {
		t.FailNow()
	}

	certificatePack, err := certificateGenerator.Generate()
	if err != nil {
		t.FailNow()
	}

	for _, file := range []string{
		certificatePack.CACertificateFilePath,
		certificatePack.CAKeyFilePath,
		certificatePack.ServerCertificateFilePath,
		certificatePack.ServerPrivateKeyFilePath,
		certificatePack.ClientCertificateFilePath,
		certificatePack.ClientPrivateKeyFilePath,
	} {
		_, err = os.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			t.FailNow()
		}
	}

	certificateBytes, err := os.ReadFile(certificatePack.ServerCertificateFilePath)
	if err != nil {
		t.FailNow()
	}

	block, _ := pem.Decode(certificateBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		t.FailNow()
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.FailNow()
	}

	for _, dnsName := range cert.DNSNames {
		if dnsName == "example.com" {
			return
		}
	}

	t.FailNow()
}
