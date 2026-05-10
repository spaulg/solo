package container

type ServiceStatus struct {
	RunningServices    []string
	StoppedServices    []string
	ExitedServices     []string
	AbsentServices     []string
	NotRunningServices []string
}
