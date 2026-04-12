package grpc

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cli_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	domain_config "github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/app/events"
	"github.com/spaulg/solo/test/mocks/host/app/wms"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
	"github.com/spaulg/solo/test/mocks/host/infra/certificate"
	"github.com/spaulg/solo/test/mocks/logging"
)

func TestMutualTLSServerFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(MutualTLSServerFactoryTestSuite))
}

type MutualTLSServerFactoryTestSuite struct {
	suite.Suite

	soloCtx                  *cli_context.CliContext
	mockProject              *project.MockProject
	mockLogHandler           *logging.MockHandler
	mockEventManager         *events.MockEventManager
	mockCertificateAuthority *certificate.MockAuthority
	mockWorkflowRunner       *wms.MockWorkflowRunner
}

func (t *MutualTLSServerFactoryTestSuite) SetupTest() {
	t.mockProject = &project.MockProject{}
	t.mockEventManager = &events.MockEventManager{}
	t.mockCertificateAuthority = &certificate.MockAuthority{}
	t.mockWorkflowRunner = &wms.MockWorkflowRunner{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.soloCtx = &cli_context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &domain.Config{
			Entrypoint: domain_config.EntrypointConfig{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			Workflow: domain_config.WorkflowConfig{
				Grpc: domain_config.GrpcConfig{
					ServerPort: 0,
				},
			},
		},
	}
}

func (t *MutualTLSServerFactoryTestSuite) TestNewMutualTLSServerFactory() {
	serverFactory := NewMutualTLSServerFactory(
		t.soloCtx,
		t.mockEventManager,
		t.mockCertificateAuthority,
		t.mockWorkflowRunner,
	)

	t.NotNil(serverFactory)
}
