package progress

type ComposeProgressEvent struct {
	Action            ComposeProgressAction
	Type              ComposeProgressEntityTypeName
	FullEntityName    string
	ProjectEntityName string
	Status            ComposeProgressStatus
}
