package progress

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/event_manager/events"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type ProgressEventPublisher struct {
	logger       *slog.Logger
	config       *domain.Config
	project      domain.Project
	eventManager events.Manager
	projectName  string
	stream       io.ReadCloser
}

func NewProgressEventPublisher(
	logger *slog.Logger,
	config *domain.Config,
	project domain.Project,
	eventManager events.Manager,
	projectName string,
	stream io.ReadCloser,
) *ProgressEventPublisher {
	return &ProgressEventPublisher{
		logger:       logger,
		config:       config,
		project:      project,
		eventManager: eventManager,
		projectName:  projectName,
		stream:       stream,
	}
}

func (t *ProgressEventPublisher) PublishStreamedProgressEvents() {
	scanner := bufio.NewScanner(t.stream)

	for scanner.Scan() {
		line := scanner.Text()

		composeProgress := ComposeProgress{}
		if err := json.Unmarshal([]byte(line), &composeProgress); err != nil {
			t.logger.Error(fmt.Sprintf("Error unmarshaling JSON: %s: %v", line, err))
			continue
		}

		if event := composeProgress.ToEvent(t.projectName); event != nil {
			t.eventManager.Publish(event)
		}
	}

	if err := scanner.Err(); err != nil {
		t.logger.Error(fmt.Sprintf("Error scanning progress stream: %v", err))
		return
	}
}
