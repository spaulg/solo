package grpc

import (
	"errors"
	"testing"
)

// error validation
//   empty hostname
//   invalid hostname
//   empty certificate base path
//   invalid certificate base path

func TestEmptyHostname(t *testing.T) {
	_, err := NewCertificateGenerator("", "")
	if err == nil {
		// fail
	}

	if !errors.Is(err, InvalidHostname) {
		// fail
	}

	// pass
}

// happy path tests
//   the certificate files are generated in the base directory
//   the server certificate is for the hostname passed
