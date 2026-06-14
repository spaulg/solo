package grpc

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/compose"
	"github.com/spaulg/solo/test/mocks/host/infra/certificate"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestMutualTLSServerFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(MutualTLSServerFactoryTestSuite))
}

type MutualTLSServerFactoryTestSuite struct {
	suite.Suite

	mockLogger               *slog.Logger
	mockProject              *compose.MockProject
	mockLogHandler           *logging.MockHandler
	mockCertificateAuthority *certificate.MockAuthority
	mockWorkflowRunner       *wms.MockWorkflowRunner
}

func (t *MutualTLSServerFactoryTestSuite) SetupTest() {
	t.mockProject = &compose.MockProject{}
	t.mockCertificateAuthority = &certificate.MockAuthority{}
	t.mockWorkflowRunner = &wms.MockWorkflowRunner{}

	t.mockLogger = slog.New(t.mockLogHandler)

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)
}

func (t *MutualTLSServerFactoryTestSuite) TestNewMutualTLSServerFactory() {
	serverFactory := NewMutualTLSServerFactory(
		t.mockLogger,
		t.mockProject,
		t.mockCertificateAuthority,
		t.mockWorkflowRunner,
	)

	t.NotNil(serverFactory)
}
