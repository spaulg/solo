package wms

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
)

type WorkflowGuard struct {
	soloCtx                   *context.CliContext
	channelLock               sync.Mutex
	containers                []string
	workflowContainerChannels map[workflowcommon.WorkflowName]map[string]chan int
	workflowContainerStatus   map[workflowcommon.WorkflowName]map[string]bool
}

func NewWorkflowGuard(
	soloCtx *context.CliContext,
	workflows []workflowcommon.WorkflowName,
	containers []string,
) wms_types.WorkflowGuard {
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
		soloCtx:                   soloCtx,
		channelLock:               sync.Mutex{},
		containers:                containers,
		workflowContainerChannels: workflowContainerChannels,
		workflowContainerStatus:   workflowContainerStatus,
	}
}

func (t *WorkflowGuard) Publish(event events_types.Event) {
	var workflowName workflowcommon.WorkflowName
	var containerName string
	var workflowSuccessful = true

	switch e := event.(type) {
	case *wms_types.WorkflowSkippedEvent:
		workflowName = e.WorkflowName
		containerName = e.FullContainerName

		t.soloCtx.Logger.Debug(fmt.Sprintf("Received event skipped for workflow %s for container %s", workflowName, containerName))

	case *wms_types.WorkflowCompleteEvent:
		workflowName = e.WorkflowName
		containerName = e.FullContainerName

		t.soloCtx.Logger.Debug(fmt.Sprintf("Received event completed for workflow %s for container %s", workflowName, containerName))
		workflowSuccessful = e.Successful

	default:
		return
	}

	if _, ok := t.workflowContainerChannels[workflowName]; !ok {
		t.soloCtx.Logger.Warn(fmt.Sprintf("Workflow %s not registered for workflow guard", workflowName))
		return
	}

	t.channelLock.Lock()
	defer t.channelLock.Unlock()

	if _, ok := t.workflowContainerChannels[workflowName][containerName]; !ok {
		t.soloCtx.Logger.Warn(fmt.Sprintf("Container %s not registered for workflow guard or channel already closed", containerName))
		return
	}

	if t.workflowContainerChannels[workflowName][containerName] == nil {
		t.soloCtx.Logger.Warn(fmt.Sprintf("Channel for container %s and workflow %s is nil, cannot close channel", containerName, workflowName))
		return
	}

	t.soloCtx.Logger.Debug(fmt.Sprintf("Closing channel for workflow %s and container %s", workflowName, containerName))

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
					t.soloCtx.Logger.Info(fmt.Sprintf("Cannot wait for workflow %s to complete as this is not mapped", workflowName))
					return fmt.Errorf("unrecognised workflow %s", workflowName)
				}

				var stopped int32

				duration := t.soloCtx.Project.GetMaxWorkflowTimeout(workflowName.String())
				timer := time.NewTimer(duration)
				startTime := time.Now()

				t.soloCtx.Logger.Info(fmt.Sprintf("Waiting for %s to complete %s workflow, time remaining: %v", container, workflowName, duration.Seconds()))

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
							t.soloCtx.Logger.Info(fmt.Sprintf("Waiting for %s to complete %s workflow, time remaining: %v", container, workflowName, remainingRounded))
						}
					}
				}()

				// Wait for confirmation the container provisioning
				// or expiry of the timer
				select {
				case <-timer.C:
					t.soloCtx.Logger.Error(fmt.Sprintf("%s failed to complete workflow %s before timeout", container, workflowName))
					return errors.New("provisioning timer expired")

				case <-t.workflowContainerChannels[workflowName][container]:
					t.soloCtx.Logger.Info(fmt.Sprintf("%s completed workflow %s before timeout", container, workflowName))
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
				t.soloCtx.Logger.Error(fmt.Sprintf("Error waiting for container %s: %v", container, err))

				lock.Lock()
				errs = append(errs, err)
				lock.Unlock()
			} else {
				t.soloCtx.Logger.Info(fmt.Sprintf("Container %s completed successfully", container))
			}

			wg.Done()
		}(container)
	}

	wg.Wait()

	if len(errs) > 0 {
		t.soloCtx.Logger.Error(fmt.Sprintf("Encountered %d errors while waiting for containers: %v", len(errs), errs))
		return fmt.Errorf("encountered errors while waiting for containers: %v", errs)
	}

	t.soloCtx.Logger.Info("All containers completed successfully")
	return nil
}
