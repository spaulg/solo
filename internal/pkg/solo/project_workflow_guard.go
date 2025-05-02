package solo

import (
	"errors"
	"fmt"
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
	"math"
	"time"
)

type WorkflowServiceMap map[workflowcommon.WorkflowName][]string
type WorkflowChannelMap map[workflowcommon.WorkflowName]chan interface{}

type ProjectWorkflowGuard struct {
	soloCtx          *context.CliContext
	workflowServices WorkflowServiceMap
	workflowStatus   WorkflowServiceMap
	workflowComplete WorkflowChannelMap
}

func NewProjectWorkflowGuard(soloCtx *context.CliContext) *ProjectWorkflowGuard {
	return &ProjectWorkflowGuard{
		soloCtx:          soloCtx,
		workflowServices: make(WorkflowServiceMap),
		workflowStatus:   make(WorkflowServiceMap),
		workflowComplete: make(WorkflowChannelMap),
	}
}

func (t *ProjectWorkflowGuard) AddWorkflow(workflow workflowcommon.WorkflowName, services []string) {
	t.workflowServices[workflow] = services
	t.workflowComplete[workflow] = make(chan interface{})
}

func (t *ProjectWorkflowGuard) Publish(event events.Event) {
	var workflowName workflowcommon.WorkflowName

	switch e := event.(type) {
	case *wms.WorkflowCompleteEvent:
		workflowName = e.WorkflowName
		t.workflowStatus[workflowName] = append(t.workflowStatus[workflowName], e.ServiceName)
		t.soloCtx.Logger.Debug(fmt.Sprintf("Received event %s for service %s", workflowName, e.ServiceName))
	default:
		return
	}

	if len(t.workflowServices[workflowName]) == len(t.workflowStatus[workflowName]) {
		t.soloCtx.Logger.Debug(fmt.Sprintf("All services completed workflow %s. Closing channel", workflowName))
		close(t.workflowComplete[workflowName])
	}
}

func (t *ProjectWorkflowGuard) WaitForCompletion(workflowName workflowcommon.WorkflowName) error {
	duration := t.soloCtx.Project.GetMaxWorkflowTimeout(workflowName.String())
	timer := time.NewTimer(duration)
	startTime := time.Now()
	stopped := false

	// If the workflow is not present in the map, return immediately
	if _, ok := t.workflowComplete[workflowName]; !ok {
		t.soloCtx.Logger.Info(fmt.Sprintf("Cannot wait for workflow %s to complete as this is not mapped", workflowName))
		return nil
	}

	t.soloCtx.Logger.Info(fmt.Sprintf("Waiting for services to complete %s workflow, time remaining: %v", workflowName, duration.Seconds()))

	// Report timer status through logs
	go func() {
		for {
			time.Sleep(1 * time.Second)
			remainingDuration := duration - time.Since(startTime)
			remaining := remainingDuration.Seconds()

			if stopped || remaining < 0 {
				return
			}

			remainingRounded := int(math.Floor(remaining))
			if (remainingRounded % 10) == 0 {
				t.soloCtx.Logger.Info(fmt.Sprintf("Waiting for services to complete %s workflow, time remaining: %v", workflowName, remainingRounded))
			}
		}
	}()

	// Wait for confirmation all containers have provisioned
	// or expiry of the timer
	select {
	case <-timer.C:
		t.soloCtx.Logger.Error(fmt.Sprintf("One or more services failed to complete workflow %s before timeout", workflowName))
		return errors.New("provisioning timer expired")

	case <-t.workflowComplete[workflowName]:
		t.soloCtx.Logger.Info(fmt.Sprintf("All services completed workflow %s before timeout", workflowName))
		timer.Stop()
		stopped = true
		return nil
	}
}
