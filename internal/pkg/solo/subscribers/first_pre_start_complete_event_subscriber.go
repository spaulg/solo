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

type FirstPreStartCompleteEventSubscriber struct {
	soloCtx *context.CliContext
}

func NewFirstPreStartCompleteEventSubscriber(soloCtx *context.CliContext) events.Subscriber {
	return &FirstPreStartCompleteEventSubscriber{
		soloCtx: soloCtx,
	}
}

func (t *FirstPreStartCompleteEventSubscriber) Publish(event events.Event) {
	switch e := event.(type) {
	case *wms.WorkflowCompleteEvent:
		if e.WorkflowName == commonworkflow.FirstPreStart && e.Successful {
			t.soloCtx.Logger.Info(fmt.Sprintf("Writing first_pre_start complete marker for %s", e.ServiceName))

			markerFile := path.Join(t.soloCtx.Project.GetServiceMountDirectory(e.ServiceName), "first_pre_start_complete")
			if _, err := os.Create(markerFile); err != nil {
				t.soloCtx.Logger.Error("Failed to write first_pre_start marker completion file: %s: %v", markerFile, err)
			}
		}
	}
}
