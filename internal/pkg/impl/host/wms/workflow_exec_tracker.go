package wms

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
)

type WorkflowExecTracker struct {
	mu           sync.Mutex
	workflowMap  map[string]struct{}
	workflowList []string
	filePath     string
}

func LoadWorkflowExecTracker(filePath string) (*WorkflowExecTracker, error) {
	tracker := &WorkflowExecTracker{
		workflowMap:  make(map[string]struct{}),
		workflowList: make([]string, 0),
		filePath:     filePath,
	}

	if _, err := os.Stat(tracker.filePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		return tracker, nil
	}

	// Load existing workflowMap from file if it exists
	data, err := os.ReadFile(tracker.filePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &tracker.workflowList); err != nil {
		return nil, err
	}

	for _, key := range tracker.workflowList {
		tracker.workflowMap[key] = struct{}{}
	}

	return tracker, nil
}

func (t *WorkflowExecTracker) Save() error {
	// Marshal to JSON
	data, err := json.Marshal(t.workflowList)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow list: %w", err)
	}

	// Write to file
	if err := os.WriteFile(t.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write workflow list to file: %w", err)
	}

	return nil
}

func (t *WorkflowExecTracker) MarkExecuted(serviceName string, workflowName commonworkflow.WorkflowName) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := serviceName + ":" + workflowName.String()
	_, loaded := t.workflowMap[key]
	if !loaded {
		t.workflowMap[key] = struct{}{}
		t.workflowList = append(t.workflowList, key)
	}

	// Save workflowMap after modification
	if err := t.Save(); err != nil {
		return !loaded, fmt.Errorf("marked executed but failed to save workflow list: %w", err)
	}

	return !loaded, nil
}

func (t *WorkflowExecTracker) Clear(serviceName []string, workflowNames []commonworkflow.WorkflowName) error {
	if len(serviceName) > 0 && len(workflowNames) > 0 {
		t.mu.Lock()
		defer t.mu.Unlock()

		// Specific services
		for _, serviceName := range serviceName {
			for _, workflowName := range workflowNames {
				key := serviceName + ":" + workflowName.String()
				delete(t.workflowMap, key)
			}
		}

		t.workflowList = make([]string, 0)
		for key := range t.workflowMap {
			t.workflowList = append(t.workflowList, key)
		}

		if err := t.Save(); err != nil {
			return fmt.Errorf("failed to clear workflow(s) and save workflow list: %w", err)
		}
	}

	return nil
}
