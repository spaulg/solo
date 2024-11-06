package grpc

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

func TestInvalidCertificateBasePath(t *testing.T) {
	_, err := NewCertificateGenerator("example.com", "/foo/bar/baz")
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

	if err = certificateGenerator.Generate(); err != nil {
		t.FailNow()
	}

	for _, file := range []string{
		certificateGenerator.CACertificateFilePath,
		certificateGenerator.CAKeyFilePath,
		certificateGenerator.ServerCertificateFilePath,
		certificateGenerator.ServerPrivateKeyFilePath,
		certificateGenerator.ClientCertificateFilePath,
		certificateGenerator.ClientPrivateKeyFilePath,
	} {
		_, err = os.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			t.FailNow()
		}
	}

	certificateBytes, err := os.ReadFile(certificateGenerator.ServerCertificateFilePath)
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
