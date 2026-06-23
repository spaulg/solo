package wfsession

type RunCommandRequest struct {
	Command          string
	Arguments        []string
	WorkingDirectory string
}
