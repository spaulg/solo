package project_file

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
)

type ProjectFile struct {
	FilePath string
}

func New(projectFilePath string) *ProjectFile {
	return &ProjectFile{
		FilePath: projectFilePath,
	}
}

// ExportComposeConfiguration takes a project file and exports a valid compose file,
// decorated with the necessary config for starting the project
func (d *ProjectFile) ExportComposeConfiguration() ([]byte, error) {
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

	for index, service := range project.Services {
		// todo: override any user config to apply root
		service.User = "0"

		// todo: replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/solo-entrypoint.sh"}

		// todo: append volume mounts for the new entrypoint, build scripts, run scripts and preferred user id config
		service.Volumes = append(service.Volumes, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   "./prototype/solo-entrypoint.sh",
			Target:   "/solo-entrypoint.sh",
			ReadOnly: true,
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   "/Users/spaulg/Repositories/spaulg/solo/prototype/build-scripts",
			Target:   "/build-scripts",
			ReadOnly: true,
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   "/Users/spaulg/Repositories/spaulg/solo/prototype/run-scripts",
			Target:   "/run-scripts",
			ReadOnly: true,
		})

		project.Services[index] = service
	}

	return project.MarshalYAML()
}
