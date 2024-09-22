package project_file

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
)

type ProjectFile struct {
	FilePath string
}

func New(projectFilePath string) *ProjectFile {
	return &ProjectFile{
		FilePath: projectFilePath,
	}
}

// GenerateCompose takes a project file and exports a valid compose file,
// decorated with the necessary config for starting the project
func (d *ProjectFile) GenerateCompose() []byte {
	projectOptionsLoader := cli.WithLoadOptions(func(option *loader.Options) {
		option.SkipValidation = true
	})

	projectOptions, err := cli.NewProjectOptions([]string{d.FilePath}, projectOptionsLoader)

	if err != nil {
		fmt.Println(fmt.Errorf("error building project options: %v", err))
	}

	project, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		fmt.Println(fmt.Errorf("error loading project: %v", err))
	}

	// todo: replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
	// todo: override any user config to apply root
	// todo: append volume mounts for the new entrypoint, build scripts, run scripts and preferred user id config

	//for _, service := range project.Services {
	//	for k := range service.Volumes {
	//		fmt.Println(k)
	//	}
	//}
	//
	//fmt.Println("Project volumes: ")
	//for k := range project.Volumes {
	//	fmt.Println(k)
	//}
	//
	//for name := range project.VolumeNames() {
	//	fmt.Println(name)
	//}

	// Write output to a file for use with docker compose command call
	bytes, err := project.MarshalYAML()
	return bytes
}
