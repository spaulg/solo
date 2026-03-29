package progress

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
)

type ProgressEventStreamer struct {
	soloCtx      *context.CliContext
	eventManager events_types.Manager
	projectName  string
	stream       io.ReadCloser
}

func NewProgressEventPublisher(
	soloCtx *context.CliContext,
	eventManager events_types.Manager,
	projectName string,
	stream io.ReadCloser,
) *ProgressEventStreamer {
	return &ProgressEventStreamer{
		soloCtx:      soloCtx,
		eventManager: eventManager,
		projectName:  projectName,
		stream:       stream,
	}
}

func (t *ProgressEventStreamer) PublishStreamedProgressEvents() {
	scanner := bufio.NewScanner(t.stream)

	for scanner.Scan() {
		line := scanner.Text()

		composeProgress := ComposeProgress{}
		if err := json.Unmarshal([]byte(line), &composeProgress); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf("Error unmarshaling JSON: %s: %v", line, err))
			continue
		}

		if event := composeProgress.ToEvent(t.projectName); event != nil {
			t.eventManager.Publish(event)
		}
	}

	if err := scanner.Err(); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf("Error scanning progress stream: %v", err))
		return
	}
}
