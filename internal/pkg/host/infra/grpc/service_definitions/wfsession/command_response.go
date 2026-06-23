package wfsession

type CommandResponse struct {
	Stdout   string
	Stderr   string
	ExitCode *uint8
}
