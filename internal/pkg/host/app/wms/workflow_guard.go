package wms

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/domain/events"
)

type WorkflowGuard struct {
	logger                    *slog.Logger
	config                    *domain2.Config
	project                   domain2.Project
	channelLock               sync.Mutex
	containers                []string
	workflowContainerChannels map[workflowcommon.WorkflowName]map[string]chan int
	workflowContainerStatus   map[workflowcommon.WorkflowName]map[string]bool
}

func NewWorkflowGuard(
	logger *slog.Logger,
	config *domain2.Config,
	project domain2.Project,
	workflows []workflowcommon.WorkflowName,
	containers []string,
) *WorkflowGuard {
	workflowContainerChannels := make(map[workflowcommon.WorkflowName]map[string]chan int)
	workflowContainerStatus := make(map[workflowcommon.WorkflowName]map[string]bool)

	for _, workflow := range workflows {
		workflowContainerChannels[workflow] = make(map[string]chan int)
		workflowContainerStatus[workflow] = make(map[string]bool)

		for _, container := range containers {
			workflowContainerChannels[workflow][container] = make(chan int)
			workflowContainerStatus[workflow][container] = true
		}
	}

	return &WorkflowGuard{
		logger:                    logger,
		config:                    config,
		project:                   project,
		channelLock:               sync.Mutex{},
		containers:                containers,
		workflowContainerChannels: workflowContainerChannels,
		workflowContainerStatus:   workflowContainerStatus,
	}
}

func (t *WorkflowGuard) Publish(event events.Event) {
	var workflowName workflowcommon.WorkflowName
	var containerName string
	var workflowSuccessful = true

	switch e := event.(type) {
	case *wf.SkippedEvent:
		workflowName = e.WorkflowName
		containerName = e.FullContainerName

		t.logger.Debug("Received event skipped", slog.String("workflow", workflowName.String()), slog.String("container", containerName))

	case *wf.CompleteEvent:
		workflowName = e.WorkflowName
		containerName = e.FullContainerName

		t.logger.Debug("Received event completed", slog.String("workflow", workflowName.String()), slog.String("container", containerName))
		workflowSuccessful = e.Successful

	default:
		return
	}

	if _, ok := t.workflowContainerChannels[workflowName]; !ok {
		t.logger.Warn("Workflow not registered for workflow guard", slog.String("workflow", workflowName.String()))
		return
	}

	t.channelLock.Lock()
	defer t.channelLock.Unlock()

	if _, ok := t.workflowContainerChannels[workflowName][containerName]; !ok {
		t.logger.Warn("Container not registered for workflow guard or channel already closed", slog.String("container", containerName))
		return
	}

	if t.workflowContainerChannels[workflowName][containerName] == nil {
		t.logger.Warn("Channel is nil, cannot close channel", slog.String("container", containerName), slog.String("workflow", workflowName.String()))
		return
	}

	t.logger.Debug("Closing channel", slog.String("workflow", workflowName.String()), slog.String("container", containerName))

	currentStatus := t.workflowContainerStatus[workflowName][containerName]
	t.workflowContainerStatus[workflowName][containerName] = currentStatus && workflowSuccessful

	close(t.workflowContainerChannels[workflowName][containerName])
}

func (t *WorkflowGuard) Wait(callback func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error) error {
	wg := sync.WaitGroup{}
	wg.Add(len(t.containers))

	lock := sync.Mutex{}
	var errs []error

	for _, container := range t.containers {
		go func(container string) {
			err := callback(container, func(workflowName workflowcommon.WorkflowName) error {
				// If the workflow is not present in the map, return immediately
				if _, ok := t.workflowContainerChannels[workflowName]; !ok {
					t.logger.Info("Cannot wait for workflow to complete as this is not mapped", slog.String("workflow", workflowName.String()))
					return fmt.Errorf("unrecognised workflow %s", workflowName)
				}

				var stopped int32

				duration := t.project.GetMaxWorkflowTimeout(workflowName.String())
				timer := time.NewTimer(duration)
				startTime := time.Now()

				t.logger.Info("Waiting for container to complete workflow", slog.String("container", container), slog.String("workflow", workflowName.String()), slog.Float64("remaining_seconds", duration.Seconds()))

				// Report timer status through logs
				go func() {
					for {
						time.Sleep(1 * time.Second)
						remainingDuration := duration - time.Since(startTime)
						remaining := remainingDuration.Seconds()

						if atomic.LoadInt32(&stopped) == 1 || remaining < 0 {
							return
						}

						remainingRounded := int(math.Floor(remaining))
						if (remainingRounded % 10) == 0 {
							t.logger.Info("Waiting for container to complete workflow", slog.String("container", container), slog.String("workflow", workflowName.String()), slog.Int("remaining_seconds", remainingRounded))
						}
					}
				}()

				// Wait for confirmation the container provisioning
				// or expiry of the timer
				select {
				case <-timer.C:
					t.logger.Error("Container failed to complete workflow before timeout", slog.String("container", container), slog.String("workflow", workflowName.String()))
					return errors.New("provisioning timer expired")

				case <-t.workflowContainerChannels[workflowName][container]:
					t.logger.Info("Container completed workflow before timeout", slog.String("container", container), slog.String("workflow", workflowName.String()))
					timer.Stop()
					atomic.StoreInt32(&stopped, 1)

					status, ok := t.workflowContainerStatus[workflowName][container]
					if !ok || !status {
						return fmt.Errorf("workflow %s for container %s did not complete successfully", workflowName, container)
					}

					return nil
				}
			})

			if err != nil {
				t.logger.Error("Error waiting for container", slog.String("container", container), slog.Any("error", err))

				lock.Lock()
				errs = append(errs, err)
				lock.Unlock()
			} else {
				t.logger.Info("Container completed successfully", slog.String("container", container))
			}

			wg.Done()
		}(container)
	}

	wg.Wait()

	if len(errs) > 0 {
		t.logger.Error("Encountered errors while waiting for containers", slog.Int("count", len(errs)), slog.Any("errors", errs))
		return fmt.Errorf("encountered errors while waiting for containers: %v", errs)
	}

	t.logger.Info("All containers completed successfully")
	return nil
}
