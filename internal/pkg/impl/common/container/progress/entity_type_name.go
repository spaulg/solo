package progress

type ProgressEntityTypeName int

const (
	UnknownEntityType ProgressEntityTypeName = iota
	Volume
	Network
	Container
	Image
)

func (t ProgressEntityTypeName) String() string {
	switch t {
	case Volume:
		return "Volume"
	case Network:
		return "Network"
	case Container:
		return "Container"
	case Image:
		return "Image"
	default:
		return "Unknown"
	}
}

func EntityTypeNameFromString(entityTypeName string) ProgressEntityTypeName {
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
