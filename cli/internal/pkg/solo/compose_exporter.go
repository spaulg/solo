package solo

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"path"
)

// ExportComposeConfiguration takes a project file and exports a valid compose file,
// decorated with the necessary config for starting the project
func ExportComposeConfiguration(globalConfig *Config, d *ProjectFile) ([]byte, error) {
	projectOptionsLoader := cli.WithLoadOptions(func(option *loader.Options) {
		option.SkipValidation = true // Prevent validation failures from preventing the global config from being loaded
		option.ResolvePaths = false  // Keep paths relative in case the user moves their project folder
	})

	projectOptions, err := cli.NewProjectOptions([]string{d.FilePath}, projectOptionsLoader)

	if err != nil {
		fmt.Println(fmt.Errorf("error building project options: %v", err))
	}

	soloEntrypoint := globalConfig.Entrypoint
	localDirectory := globalConfig.LocalDirectory

	project, err := projectOptions.LoadProject(context.Background())
	if err != nil {
		fmt.Println(fmt.Errorf("error loading project: %v", err))
	}

	for index, service := range project.Services {
		// Override any user globalConfig to force root
		// todo: allow the user to switch in the entrypoint script
		service.User = "0"

		// Replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
		if len(service.Entrypoint) > 0 {
			service.Command = append(service.Entrypoint, service.Command...)
		}

		service.Entrypoint = []string{"/solo-entrypoint"}

		// Append volume mounts for the new entrypoint, build scripts, run scripts and preferred user id globalConfig
		service.Volumes = append(service.Volumes, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   soloEntrypoint,
			Target:   "/solo-entrypoint",
			ReadOnly: true,
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   path.Join(localDirectory, "build-scripts"),
			Target:   "/build-scripts",
			ReadOnly: true,
			Bind: &types.ServiceVolumeBind{
				CreateHostPath: true,
			},
		}, types.ServiceVolumeConfig{
			Type:     "bind",
			Source:   path.Join(localDirectory, "run-scripts"),
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
