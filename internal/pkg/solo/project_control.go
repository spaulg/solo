package solo

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/event"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc"
	"github.com/spaulg/solo/internal/pkg/solo/orchestrator"
	"os"
	"path"
	"time"
)

type ProjectControl struct {
	soloCtx           *context.SoloContext
	composeFile       string
	orchestrator      orchestrator.Orchestrator
	grpcServerFactory grpc.ServerFactory
	eventStream       event.Stream[events.ProvisioningEvent]
}

func NewProjectControl(
	soloCtx *context.SoloContext,
	orchestrator orchestrator.Orchestrator,
	grpcServerFactory grpc.ServerFactory,
	eventStream event.Stream[events.ProvisioningEvent],
) *ProjectControl {
	return &ProjectControl{
		soloCtx:           soloCtx,
		composeFile:       soloCtx.Project.ResolveStateDirectory("docker-compose.yml"),
		orchestrator:      orchestrator,
		grpcServerFactory: grpcServerFactory,
		eventStream:       eventStream,
	}
}

func (t *ProjectControl) Start() error {
	grpcServer, err := t.grpcServerFactory.Build(
		t.orchestrator.GetHostGatewayHostname(),
		t.soloCtx.Config.GrpcServerPort,
		t.soloCtx.Project,
	)

	if err != nil {
		return err
	}

	// Start GRPC services
	if err := grpcServer.Start(); err != nil {
		return err
	}

	defer grpcServer.Stop()

	// Write compose file
	composeYml, _ := t.orchestrator.ExportComposeConfiguration(t.soloCtx.Config, t.soloCtx.Project)
	if err := t.exportComposeFile(composeYml); err != nil {
		return err
	}

	// Start compose services
	if err := t.orchestrator.Up(t.soloCtx.Project.GetDirectory(), t.composeFile); err != nil {
		return fmt.Errorf("error running composeCmd: %v", err)
	}

	duration := 30 * time.Second
	timer := time.NewTimer(duration)
	startTime := time.Now()

	interrupt := make(chan struct{})

	// Go routine to report timer status
	go func() {
		for {
			time.Sleep(1 * time.Second)
			remaining := duration - time.Since(startTime)
			if remaining <= 0 {
				return
			}

			fmt.Printf("Time remaining: %v\n", remaining)
		}
	}()

	// Wait for confirmation all containers have provisioned
	// or expiry of the timer
	select {
	case <-timer.C:
		fmt.Println("Timer expired")
		return errors.New("provisioning timer expired")
	case <-interrupt:
		fmt.Println("All containers reported finished")
	}

	// todo: Exec post start commands (via docker exec)
	// todo: wait delay period for all containers to checkin for post start commands provisioning

	return nil
}

func (t *ProjectControl) Stop() error {
	if err := t.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := t.orchestrator.Down(t.soloCtx.Project.GetDirectory(), t.composeFile); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	return nil
}

func (t *ProjectControl) Destroy(purgeStateDirectory bool) error {
	if err := t.composeFileExists(); err != nil {
		return err
	}

	// todo: Exec pre stop commands

	if err := t.orchestrator.Destroy(t.soloCtx.Project.GetDirectory(), t.composeFile); err != nil {
		return fmt.Errorf("error running compose: %v", err)
	}

	var purgeDirectoryList []string

	if purgeStateDirectory {
		// Purge the entire state directory
		purgeDirectoryList = []string{t.soloCtx.Project.GetStateDirectoryRoot()}
	} else {
		// Purge only certain sub folders
		purgeDirectoryList = []string{
			t.soloCtx.Project.GetAllServicesStateDirectory(),
			t.soloCtx.Project.GetServiceStateDirectoryRoot(),
		}
	}

	for _, purgeDirectory := range purgeDirectoryList {
		if err := os.RemoveAll(purgeDirectory); err != nil {
			return fmt.Errorf("failed to remove state directory %s: %v", purgeDirectory, err)
		}
	}

	return nil
}

func (t *ProjectControl) exportComposeFile(composeYml []byte) error {
	composeDirectory := path.Dir(t.composeFile)
	if _, err := os.Stat(composeDirectory); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check .solo directory existence: %v", err)
		}

		if err := os.MkdirAll(composeDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create .solo directory: %v", err)
		}
	}

	if err := os.WriteFile(t.composeFile, composeYml, 0640); err != nil {
		return fmt.Errorf("failed to write compose file: %v", err)
	}

	return nil
}

func (t *ProjectControl) composeFileExists() error {
	if _, err := os.Stat(t.composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("compose file not found")
		} else {
			return fmt.Errorf("error looking for compose file: %v", err)
		}
	}

	return nil
}
