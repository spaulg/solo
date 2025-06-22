package certificate

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestTestCertificateAuthority(t *testing.T) {
	suite.Run(t, new(TestCertificateAuthority))
}
