package project_file

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/spf13/viper"
	"path"
	"path/filepath"
	"strings"
)

type ProjectFile struct {
	Directory string
	FilePath  string
}

func New(projectFilePath string) *ProjectFile {
	return &ProjectFile{
		Directory: filepath.Dir(projectFilePath),
		FilePath:  projectFilePath,
	}
}

// ExportComposeConfiguration takes a project file and exports a valid compose file,
// decorated with the necessary config for starting the project
func (d *ProjectFile) ExportComposeConfiguration(config *viper.Viper) ([]byte, error) {
	projectOptionsLoader := cli.WithLoadOptions(func(option *loader.Options) {
		option.SkipValidation = true // Prevent validation failures from preventing the config from being loaded
		option.ResolvePaths = false  // Keep paths relative in case the user moves their project folder
	})

	projectOptions, err := cli.NewProjectOptions([]string{d.FilePath}, projectOptionsLoader)

	if err != nil {
		fmt.Println(fmt.Errorf("error building project options: %v", err))
	}

	soloEntrypoint := config.GetString("Entrypoint")
	localDirectory := config.GetString("LocalDirectory")

	project, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		fmt.Println(fmt.Errorf("error loading project: %v", err))
	}

	for index, service := range project.Services {
		// Override any user config to force root
		// todo: allow the user to switch in the entrypoint script
		service.User = "0"

		// Replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/solo-entrypoint.sh"}

		// Patch relative paths to account for the project work directory
		for index, volume := range service.Volumes {
			if volume.Type == "bind" && !strings.HasPrefix(volume.Source, "/") {
				service.Volumes[index].Source = "../" + volume.Source
			}
		}

		// Append volume mounts for the new entrypoint, build scripts, run scripts and preferred user id config
		service.Volumes = append(service.Volumes, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   soloEntrypoint,
			Target:   "/solo-entrypoint.sh",
			ReadOnly: true,
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   path.Join("..", localDirectory, "build-scripts"),
			Target:   "/build-scripts",
			ReadOnly: true,
			Bind: &types.ServiceVolumeBind{
				CreateHostPath: true,
			},
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   path.Join("..", localDirectory, "run-scripts"),
			Target:   "/run-scripts",
			ReadOnly: true,
			Bind: &types.ServiceVolumeBind{
				CreateHostPath: true,
			},
		})

		project.Services[index] = service
	}

	return project.MarshalYAML()
}
