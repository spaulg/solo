package progress

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"io"
)

type ProgressEventStreamer struct {
	soloCtx      *context.CliContext
	eventManager events.Manager
	projectName  string
	stream       io.ReadCloser
}

func NewProgressEventPublisher(
	soloCtx *context.CliContext,
	eventManager events.Manager,
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
