package project

type Project interface {
	Start()
	Stop()
	Destroy()
	ComposeConfig()
}
