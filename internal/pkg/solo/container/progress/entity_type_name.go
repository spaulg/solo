package progress

type ComposeProgressEntityTypeName int

const (
	UnknownEntityType ComposeProgressEntityTypeName = iota
	Volume
	Network
	Container
	Image
)

func EntityTypeNameFromString(entityTypeName string) ComposeProgressEntityTypeName {
	switch entityTypeName {
	case "Volume":
		return Volume
	case "Network":
		return Network
	case "Container":
		return Container
	case "Image":
		return Image
	default:
		return UnknownEntityType
	}
}
