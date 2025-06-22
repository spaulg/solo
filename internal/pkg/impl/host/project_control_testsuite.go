package host

import (
	"errors"
	"log/slog"
	"os"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/container"
	"github.com/spaulg/solo/test/mocks/host/events"
	"github.com/spaulg/solo/test/mocks/host/grpc"
	"github.com/spaulg/solo/test/mocks/host/logging"
	"github.com/spaulg/solo/test/mocks/host/project"
	"github.com/spaulg/solo/test/mocks/host/wms"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProjectControlTestSuite struct {
	suite.Suite

	soloCtx                  *context.CliContext
	mockProject              *project.MockProject
	mockOrchestratorFactory  *container.MockOrchestratorFactory
	mockGrpcServerFactory    *grpc.MockGRPCServerFactory
	mockWorkflowManager      *events.MockEventManager
	mockOrchestrator         *container.MockOrchestrator
	mockLogHandler           *logging.MockHandler
	mockGrpcServer           *grpc.MockAsynchronousServer
	mockWorkflowGuardFactory *wms.MockWorkflowGuardFactory
	mockWorkflowGuard        *wms.MockWorkflowGuard
}

func (t *ProjectControlTestSuite) SetupTest() {
	t.mockOrchestratorFactory = &container.MockOrchestratorFactory{}
	t.mockGrpcServerFactory = &grpc.MockGRPCServerFactory{}
	t.mockWorkflowManager = &events.MockEventManager{}
	t.mockOrchestrator = &container.MockOrchestrator{}
	t.mockProject = &project.MockProject{}
	t.mockGrpcServer = &grpc.MockAsynchronousServer{}
	t.mockWorkflowGuardFactory = &wms.MockWorkflowGuardFactory{}
	t.mockWorkflowGuard = &wms.MockWorkflowGuard{}

	t.mockLogHandler = &logging.MockHandler{}
	t.mockLogHandler.On("Enabled", mock.Anything, mock.Anything).Return(true)

	t.soloCtx = &context.CliContext{
		Project: t.mockProject,
		Logger:  slog.New(t.mockLogHandler),
		Config: &config_types.Config{
			Entrypoint: config_types.Entrypoint{
				HostEntrypointPath: test.GetTestDataFilePath("entrypoint.sh"),
			},
			GrpcServerPort: 0,
		},
	}
}

func (t *ProjectControlTestSuite) TestStart_OrchestratorFactoryReturnsError() {
	t.mockOrchestratorFactory.On("Build").Return(nil, errors.New("mock orchestrator build error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to build orchestrator")
	t.ErrorContains(err, "mock orchestrator build error")
}

func (t *ProjectControlTestSuite) TestStart_ServiceStatusReturnsError() {
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(nil, errors.New("mock services status error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to check service status")
	t.ErrorContains(err, "mock services status error")
}

func (t *ProjectControlTestSuite) TestStart_AllServicesAlreadyRunning() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "All required services already running"
	})).Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.Nil(err)
}

func (t *ProjectControlTestSuite) TestStart_GRPCServerFailsToBuild() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)

	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(nil, errors.New("mock grpc server build error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to build GRPC server")
	t.ErrorContains(err, "mock grpc server build error")
}

func (t *ProjectControlTestSuite) TestStart_GRPCServerFailsToStart() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(errors.New("mock grpc server start error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to start GRPC server")
	t.ErrorContains(err, "mock grpc server start error")
}

func (t *ProjectControlTestSuite) TestStart_EntrypointCopyFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)

	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)

	t.mockProject.On("GetStateDirectoryRoot").Return(t.T().TempDir())

	t.soloCtx.Config.Entrypoint.HostEntrypointPath = test.GetTestDataFilePath("non_existent_entrypoint.sh")

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to copy entrypoint to state directory")
}

func (t *ProjectControlTestSuite) TestStart_ContainerNamesFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)

	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)

	t.mockProject.On("GetStateDirectoryRoot").Return(t.T().TempDir())

	t.soloCtx.Config.Entrypoint.HostEntrypointPath = test.GetTestDataFilePath("entrypoint.sh")

	t.mockProject.On("ContainerNames", servicesStatus.NotRunningServices).Return(nil, errors.New("mock container names error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to convert service names to container names")
	t.ErrorContains(err, "mock container names error")
}

func (t *ProjectControlTestSuite) TestStart_ComposeUpFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)

	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)

	t.mockProject.On("GetStateDirectoryRoot").Return(t.T().TempDir())

	t.soloCtx.Config.Entrypoint.HostEntrypointPath = test.GetTestDataFilePath("entrypoint.sh")

	t.mockProject.On("ContainerNames", servicesStatus.NotRunningServices).Return([]string{"test_server"}, nil)

	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)

	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)

	t.mockOrchestrator.On("ComposeUp").Return(errors.New("mock orchestrator error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "failed to start services")
	t.ErrorContains(err, "mock orchestrator error")
}

func (t *ProjectControlTestSuite) TestStart_GuardWaitFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)

	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)

	t.mockProject.On("GetStateDirectoryRoot").Return(t.T().TempDir())

	t.soloCtx.Config.Entrypoint.HostEntrypointPath = test.GetTestDataFilePath("entrypoint.sh")

	t.mockProject.On("ContainerNames", servicesStatus.NotRunningServices).Return([]string{"test_server"}, nil)

	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(errors.New("mock guard wait error"))

	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)

	t.mockOrchestrator.On("ComposeUp").Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.ErrorContains(err, "error waiting for services to complete workflows")
	t.ErrorContains(err, "mock guard wait error")
}

func (t *ProjectControlTestSuite) TestStart_Succeeds() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    make([]string, 0),
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     []string{"test_server"},
		NotRunningServices: []string{"test_server"},
	}

	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir())
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)

	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)

	t.mockProject.On("GetStateDirectoryRoot").Return(t.T().TempDir())

	t.soloCtx.Config.Entrypoint.HostEntrypointPath = test.GetTestDataFilePath("entrypoint.sh")

	t.mockProject.On("ContainerNames", servicesStatus.NotRunningServices).Return([]string{"test_server"}, nil)

	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(nil)

	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)

	t.mockOrchestrator.On("ComposeUp").Return(nil)

	t.mockWorkflowManager.On("Wait").Return(nil)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Waiting for all remaining events to be delivered"
	})).Return(nil)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Finished starting all services successfully"
	})).Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Start()

	t.Nil(err)
}

// func (t *ProjectControlTestSuite) TestStop_ComposeFileDoesNotExist() {
// 	t.mockProject.On("GetGeneratedComposeFilePath").Return(t.T().TempDir() + "/nonexistent-compose.yml")

// 	projectControl := NewProjectControl(
// 		t.soloCtx,
// 		t.mockWorkflowManager,
// 		t.mockOrchestratorFactory,
// 		t.mockGrpcServerFactory,
// 		t.mockWorkflowGuardFactory,
// 	)

// 	err := projectControl.Stop()
// 	t.ErrorContains(err, "compose file does not exist")
// }

func (t *ProjectControlTestSuite) TestStop_OrchestratorFactoryFails() {
	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(nil, errors.New("mock orchestrator build error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "failed to build orchestrator")
	t.ErrorContains(err, "mock orchestrator build error")
}

func (t *ProjectControlTestSuite) TestStop_ServicesStatusFails() {
	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(nil, errors.New("mock services status error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "failed to check service status")
	t.ErrorContains(err, "mock services status error")
}

func (t *ProjectControlTestSuite) TestStop_GrpcServerBuildFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)

	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(nil, errors.New("mock grpc server build error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "failed to build GRPC server")
	t.ErrorContains(err, "mock grpc server build error")
}

func (t *ProjectControlTestSuite) TestStop_GrpcServerStartFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(errors.New("mock grpc server start error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "failed to start GRPC server")
	t.ErrorContains(err, "mock grpc server start error")
}

func (t *ProjectControlTestSuite) TestStop_ContainerNamesFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return(nil, errors.New("mock container names error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "failed to convert service names to container names")
	t.ErrorContains(err, "mock container names error")
}

func (t *ProjectControlTestSuite) TestStop_GuardWaitFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return([]string{"test_server"}, nil)
	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(errors.New("mock guard wait error"))
	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)
	t.mockOrchestrator.On("ComposeStop").Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "error waiting for services to complete workflows")
	t.ErrorContains(err, "mock guard wait error")
}

func (t *ProjectControlTestSuite) TestStop_ComposeStopFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return([]string{"test_server"}, nil)
	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(nil)
	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)
	t.mockOrchestrator.On("ComposeStop").Return(errors.New("mock compose stop error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.ErrorContains(err, "failed to stop services")
	t.ErrorContains(err, "mock compose stop error")
}

func (t *ProjectControlTestSuite) TestStop_Succeeds() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    make([]string, 0),
		ExitedServices:     make([]string, 0),
		AbsentServices:     make([]string, 0),
		NotRunningServices: make([]string, 0),
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return([]string{"test_server"}, nil)
	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(nil)
	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)
	t.mockOrchestrator.On("ComposeStop").Return(nil)
	t.mockWorkflowManager.On("Wait").Return(nil)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Finished stopping all services successfully"
	})).Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Stop()

	t.Nil(err)
}

// func (t *ProjectControlTestSuite) TestDestroy_ComposeFileDoesNotExist() {
// 	t.mockProject.On("GetGeneratedComposeFilePath").Return("/tmp/nonexistent-compose.yml")

// 	projectControl := NewProjectControl(
// 		t.soloCtx,
// 		t.mockWorkflowManager,
// 		t.mockOrchestratorFactory,
// 		t.mockGrpcServerFactory,
// 		t.mockWorkflowGuardFactory,
// 	)

// 	err := projectControl.Destroy()
// 	t.ErrorContains(err, "compose file does not exist")
// }

func (t *ProjectControlTestSuite) TestDestroy_OrchestratorFactoryFails() {
	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(nil, errors.New("mock orchestrator build error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "failed to build orchestrator")
	t.ErrorContains(err, "mock orchestrator build error")
}

func (t *ProjectControlTestSuite) TestDestroy_ServicesStatusFails() {
	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(nil, errors.New("mock services status error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "failed to check service status")
	t.ErrorContains(err, "mock services status error")
}

func (t *ProjectControlTestSuite) TestDestroy_GrpcServerBuildFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    []string{},
		ExitedServices:     []string{},
		AbsentServices:     []string{},
		NotRunningServices: []string{},
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(nil, errors.New("mock grpc server build error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "failed to build GRPC server")
	t.ErrorContains(err, "mock grpc server build error")
}

func (t *ProjectControlTestSuite) TestDestroy_GrpcServerStartFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    []string{},
		ExitedServices:     []string{},
		AbsentServices:     []string{},
		NotRunningServices: []string{},
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(errors.New("mock grpc server start error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "failed to start GRPC server")
	t.ErrorContains(err, "mock grpc server start error")
}

func (t *ProjectControlTestSuite) TestDestroy_ContainerNamesFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    []string{},
		ExitedServices:     []string{},
		AbsentServices:     []string{},
		NotRunningServices: []string{},
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return(nil, errors.New("mock container names error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "failed to convert service names to container names")
	t.ErrorContains(err, "mock container names error")
}

func (t *ProjectControlTestSuite) TestDestroy_GuardWaitFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    []string{},
		ExitedServices:     []string{},
		AbsentServices:     []string{},
		NotRunningServices: []string{},
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return([]string{"test_server"}, nil)
	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(errors.New("mock guard wait error"))
	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)
	t.mockOrchestrator.On("ComposeDown").Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "error waiting for services to complete workflows")
	t.ErrorContains(err, "mock guard wait error")
}

func (t *ProjectControlTestSuite) TestDestroy_ComposeDownFails() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    []string{},
		ExitedServices:     []string{},
		AbsentServices:     []string{},
		NotRunningServices: []string{},
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return([]string{"test_server"}, nil)
	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(nil)
	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)
	t.mockOrchestrator.On("ComposeDown").Return(errors.New("mock compose down error"))

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.ErrorContains(err, "failed to destroy services")
	t.ErrorContains(err, "mock compose down error")
}

func (t *ProjectControlTestSuite) TestDestroy_Succeeds() {
	servicesStatus := &container_types.ServiceStatus{
		RunningServices:    []string{"test_server"},
		StoppedServices:    []string{},
		ExitedServices:     []string{},
		AbsentServices:     []string{},
		NotRunningServices: []string{},
	}

	tmpDir := t.T().TempDir()
	composePath := tmpDir + "/docker-compose.yml"

	if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
		t.Fail("Failed to copy compose file for test")
	}

	t.mockProject.On("GetGeneratedComposeFilePath").Return(composePath)
	t.mockOrchestratorFactory.On("Build").Return(t.mockOrchestrator, nil)
	t.mockOrchestrator.On("ServicesStatus").Return(servicesStatus, nil)
	t.mockGrpcServerFactory.On("Build", t.mockOrchestrator, t.mockProject, 0).Return(t.mockGrpcServer, nil)
	t.mockGrpcServer.On("Start").Return(nil)
	t.mockGrpcServer.On("Stop").Return(nil)
	t.mockProject.On("ContainerNames", []string{"test_server"}).Return([]string{"test_server"}, nil)
	t.mockWorkflowGuardFactory.On("Build", mock.Anything, mock.Anything).Return(t.mockWorkflowGuard)
	t.mockWorkflowGuard.On("Wait", mock.Anything).Return(nil)
	t.mockWorkflowManager.On("Subscribe", mock.Anything)
	t.mockWorkflowManager.On("Unsubscribe", mock.Anything)
	t.mockOrchestrator.On("ComposeDown").Return(nil)
	t.mockWorkflowManager.On("Wait").Return(nil)

	t.mockLogHandler.On("Handle", mock.Anything, mock.MatchedBy(func(record slog.Record) bool {
		return record.Message == "Finished destroying all services successfully"
	})).Return(nil)

	projectControl := NewProjectControl(
		t.soloCtx,
		t.mockWorkflowManager,
		t.mockOrchestratorFactory,
		t.mockGrpcServerFactory,
		t.mockWorkflowGuardFactory,
	)

	err := projectControl.Destroy()
	t.Nil(err)
}

/*
	2. **Test Start**
		- Compose file does not exist, and exportComposeFile succeeds.
		- Compose file does not exist, and exportComposeFile fails.
		- Compose file exists.

	3. **Test Stop**
		- Compose file does not exist.

	4. **Test Destroy**
		- Compose file does not exist.

	5. **Test Clean**
		- Purge state directory: os.RemoveAll fails.
		- Purge subfolders: os.RemoveAll fails for one or more.
		- Happy path: all removals succeed.

*/
