package compose

import (
	"path/filepath"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/test"
	"github.com/spaulg/solo/test/mocks/host/domain/project"
)

func TestServiceConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceConfigTestSuite))
}

type ServiceConfigTestSuite struct {
	suite.Suite

	desiredWorkingDirectory string
	testDirectory           string
	mockProject             *project.MockProject
}

func (t *ServiceConfigTestSuite) SetupTest() {
	t.testDirectory = test.GetTestDataFilePath("project/service_config/working_directory_resolution")

	t.mockProject = &project.MockProject{}
	t.mockProject.On("GetDirectory").Return(t.testDirectory)
	t.mockProject.On("GetStateDirectoryRoot").Return(filepath.Join(t.testDirectory, ".solo"))
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectoryMultipleApplicationsWithNoSharedRoot() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path1",
		Target: "/path1",
	})

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path2",
		Target: "/path2",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path1"))
	t.Equal("/path1", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path2"))
	t.Equal("/path2", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path3"))
	t.Equal("", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectoryMultipleApplicationsWithSharedRoot() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path1",
		Target: "/app/path1",
	})

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path2",
		Target: "/app/path2",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path1"))
	t.Equal("/app/path1", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path2"))
	t.Equal("/app/path2", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path3"))
	t.Equal("", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectorySingleApplicationOfMonorepo() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path1",
		Target: "/app",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path1"))
	t.Equal("/app", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path2"))
	t.Equal("", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path3"))
	t.Equal("", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectorySingleApplicationProjectRootMapping() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./",
		Target: "/app",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path1"))
	t.Equal("/app/path1", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path2"))
	t.Equal("/app/path2", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path3"))
	t.Equal("/app/path3", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectoryMultipleApplicationsWithStackedRoot() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path1",
		Target: "/app/path1",
	})

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path2",
		Target: "/app/path1/path2",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path1"))
	t.Equal("/app/path1", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path2"))
	t.Equal("/app/path1/path2", t.desiredWorkingDirectory)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path3"))
	t.Equal("", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectoryInaccessiblePaths() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./path4",
		Target: "/app/path4",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory, "path4"))
	t.Equal("", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectoryNonDirectoryPaths() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./file1.txt",
		Target: "/app/file1.txt",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory))
	t.Equal("", t.desiredWorkingDirectory)
}

func (t *ServiceConfigTestSuite) TestResolveWorkingDirectoryStateDirectoryPath() {
	serviceConfig := types.ServiceConfig{
		Volumes: make([]types.ServiceVolumeConfig, 0),
	}

	serviceConfig.Volumes = append(serviceConfig.Volumes, types.ServiceVolumeConfig{
		Type:   "bind",
		Source: "./.solo",
		Target: "/app/solostate",
	})

	serviceConfigWrapper := NewServiceConfig(t.mockProject, serviceConfig)

	t.desiredWorkingDirectory = serviceConfigWrapper.ResolveContainerWorkingDirectory(filepath.Join(t.testDirectory))
	t.Equal("", t.desiredWorkingDirectory)
}
