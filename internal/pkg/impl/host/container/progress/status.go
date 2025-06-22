package progress

import (
	"crypto/md5"
	"encoding/hex"
	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"
	"strings"
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
	entityTypeEnum := progresscommon.EntityTypeNameFromString(entityType)

	// Context id to represent the target across multiple event
	hash := md5.Sum([]byte(action.String() + ":" + entityType + ":" + projectEntityName))
	contextId := hex.EncodeToString(hash[:])

	return &ComposeProgressEvent{
		ContextId:         contextId,         // MD5 of the status, type and entity name triple
		Action:            action,            // Start, Stop, Create, Remove, Build
		EntityType:        entityTypeEnum,    // Volume, Network, Container, Image
		FullEntityName:    fullEntityName,    // Full projectEntityName name in docker being actioned
		ProjectEntityName: projectEntityName, // Project projectEntityName name in compose being actioned
		Status:            status,            // InProgress or Complete
	}
}

func (t *ComposeProgress) convertActionStatus(status string) (progresscommon.ComposeProgressAction, progresscommon.ComposeProgressStatus) {
	switch status {
	case "Built":
		return progresscommon.Build, progresscommon.Complete
	case "Creating":
		return progresscommon.Create, progresscommon.InProgress
	case "Created":
		return progresscommon.Create, progresscommon.Complete
	case "Starting":
		return progresscommon.Start, progresscommon.InProgress
	case "Started":
		return progresscommon.Start, progresscommon.Complete
	case "Stopping":
		return progresscommon.Stop, progresscommon.InProgress
	case "Stopped":
		return progresscommon.Stop, progresscommon.Complete
	case "Removing":
		return progresscommon.Remove, progresscommon.InProgress
	case "Removed":
		return progresscommon.Remove, progresscommon.Complete

	default:
		return progresscommon.Unknown, progresscommon.UnknownProgress
	}
}
