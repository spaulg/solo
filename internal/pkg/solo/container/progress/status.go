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
	var entity string

	partsLen := len(idParts)
	if partsLen == 2 {
		idType = idParts[0]
		entity = idParts[1]
	} else if partsLen == 1 && t.Status == "Built" {
		idType = "Image"
		entity = idParts[0]
	} else {
		return nil
	}

	entity = strings.TrimSpace(entity)
	entity = strings.Trim(entity, "\"")
	entity = strings.TrimPrefix(entity, projectName)
	entity = strings.Trim(entity, "_-")

	if len(entity) == 0 {
		return nil
	}

	return &ComposeProgressEvent{
		Action: t.Status,
		Type:   idType,
		Entity: entity,
	}
}
