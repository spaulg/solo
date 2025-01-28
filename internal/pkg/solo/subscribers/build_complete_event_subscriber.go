package subscribers

import (
	"fmt"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
	"os"
	"path"
)

type BuildCompleteEventSubscriber struct {
	soloCtx *context.CliContext
}

func NewBuildCompleteEventSubscriber(soloCtx *context.CliContext) events.Subscriber {
	return &BuildCompleteEventSubscriber{
		soloCtx: soloCtx,
	}
}

func (t *BuildCompleteEventSubscriber) Publish(event events.Event) {
	switch e := event.(type) {
	case *wms.WorkflowCompleteEvent:
		if e.WorkflowName == commonworkflow.FirstPreStart && e.Successful {
			t.soloCtx.Logger.Info(fmt.Sprintf("Writing build complete marker for %s", e.ServiceName))

			markerFile := path.Join(t.soloCtx.Project.GetServiceMountDirectory(e.ServiceName), "build_complete")
			if _, err := os.Create(markerFile); err != nil {
				t.soloCtx.Logger.Error("Failed to write build marker completion file: %s: %v", markerFile, err)
			}
		}
	}
}
