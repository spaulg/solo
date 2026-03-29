package project

import (
	"time"

	"github.com/compose-spec/compose-go/v2/types"

	"github.com/spaulg/solo/internal/pkg/types/host/domain/project/compose"
)

type Project interface {
	ResolveStateDirectory(relativePath string) string
	GetAllServicesStateDirectory() string
	GetServiceStateDirectoryRoot() string
	GetServiceStateDirectory(serviceName string) string
	GetServiceLogDirectory(serviceName string) string
	GetServiceMountDirectory(serviceName string) string
	GetStateDirectoryRoot() string
	GetDirectory() string
	GetFilePath() string
	GetGeneratedComposeFilePath() string
	GetMaxWorkflowTimeout(eventName string) time.Duration
	GetCompose() *types.Project
	Tools() compose.Tools
	Services() compose.Services
	Name() string
	ReloadWithAllProfilesEnabled() (Project, error)
	ReloadWithProfiles(profiles []string) error
	Profiles() []string
}
