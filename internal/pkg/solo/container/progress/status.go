package progress

import "strings"

type ComposeProgress struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (t *ComposeProgress) ToEvent(projectName string) *ComposeProgressEvent {
	if t.ID == "" || t.Status == "" {
		return nil
	}

	idParts := strings.SplitN(t.ID, " ", 2)
	var idType string
	var fullEntityName string
	var projectEntityName string

	partsLen := len(idParts)
	if partsLen == 2 {
		idType = idParts[0]
		fullEntityName = idParts[1]
	} else if partsLen == 1 && t.Status == "Built" {
		idType = "Image"
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
	idTypeEnum := EntityTypeNameFromString(idType)

	return &ComposeProgressEvent{
		Action:            action,            // Start, Stop, Create, Remove, Build
		Type:              idTypeEnum,        // Volume, Network, Container, Image
		FullEntityName:    fullEntityName,    // Full projectEntityName name in docker being actioned
		ProjectEntityName: projectEntityName, // Project projectEntityName name in compose being actioned
		Status:            status,            // InProgress or Complete
	}
}

func (t *ComposeProgress) convertActionStatus(status string) (ComposeProgressAction, ComposeProgressStatus) {
	switch status {
	case "Built":
		return Build, Complete
	case "Creating":
		return Create, InProgress
	case "Created":
		return Create, Complete
	case "Starting":
		return Start, InProgress
	case "Started":
		return Start, Complete
	case "Stopping":
		return Stop, InProgress
	case "Stopped":
		return Stop, Complete
	case "Removing":
		return Remove, InProgress
	case "Removed":
		return Remove, Complete

	default:
		return Unknown, UnknownProgress
	}
}
