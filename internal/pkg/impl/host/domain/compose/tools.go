package compose

type Tools map[string]ToolConfig

func NewTools() Tools {
	return make(Tools)
}
