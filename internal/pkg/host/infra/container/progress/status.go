package progress

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	progress2 "github.com/spaulg/solo/internal/pkg/shared/domain/container/progress"
)

type ComposeProgress struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (t *ComposeProgress) ToEvent(projectName string) *ComposeProgressEvent {
	if t.ID == "" || t.Status == "" {
		return nil
	}

	idParts := strings.SplitN(t.ID, " ", 2)
	var entityType string
	var fullEntityName string
	var projectEntityName string

	partsLen := len(idParts)
	if partsLen == 2 {
		entityType = idParts[0]
		fullEntityName = idParts[1]
	} else if partsLen == 1 && t.Status == "Built" {
		entityType = "Image"
		fullEntityName = idParts[0]
	} else {
		return nil
	}

	fullEntityName = strings.TrimSpace(fullEntityName)
	fullEntityName = strings.Trim(fullEntityName, "\"")

	projectEntityName = strings.TrimPrefix(fullEntityName, projectName)
	projectEntityName = strings.Trim(projectEntityName, "_-")

	if len(projectEntityName) == 0 {
		return nil
	}

	// Convert the status to an action and status
	action, status := t.convertActionStatus(t.Status)
	entityTypeEnum := progress2.EntityTypeNameFromString(entityType)

	// Context id to represent the target across multiple event
	hash := md5.Sum([]byte(action.String() + ":" + entityType + ":" + projectEntityName)) // nolint:gosec
	contextID := hex.EncodeToString(hash[:])

	return &ComposeProgressEvent{
		ContextID:         contextID,         // MD5 of the status, type and entity name triple
		Action:            action,            // Start, Stop, Create, Remove, Build
		EntityType:        entityTypeEnum,    // Volume, Network, Container, Image
		FullEntityName:    fullEntityName,    // Full projectEntityName name in docker being actioned
		ProjectEntityName: projectEntityName, // Project projectEntityName name in compose being actioned
		Status:            status,            // InProgress or Complete
	}
}

func (t *ComposeProgress) convertActionStatus(status string) (progress2.ComposeProgressAction, progress2.ComposeProgressStatus) {
	switch status {
	case "Built":
		return progress2.Build, progress2.Complete
	case "Creating":
		return progress2.Create, progress2.InProgress
	case "Created":
		return progress2.Create, progress2.Complete
	case "Starting":
		return progress2.Start, progress2.InProgress
	case "Started":
		return progress2.Start, progress2.Complete
	case "Stopping":
		return progress2.Stop, progress2.InProgress
	case "Stopped":
		return progress2.Stop, progress2.Complete
	case "Removing":
		return progress2.Remove, progress2.InProgress
	case "Removed":
		return progress2.Remove, progress2.Complete

	default:
		return progress2.Unknown, progress2.UnknownProgress
	}
}
