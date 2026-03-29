package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const shortSHALength = 7
const devVersion = "dev"
const unknownVersion = "unknown"

// nolint:gochecknoglobals
var (
	Version   = devVersion
	GitCommit = unknownVersion
	BuildDate = unknownVersion
)

type Info struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
}

func Get() Info {
	var goVersion string
	buildInfo, ok := debug.ReadBuildInfo()

	if ok && buildInfo != nil {
		goVersion = strings.TrimPrefix(buildInfo.GoVersion, "go")
	} else {
		goVersion = unknownVersion
	}

	return Info{
		Version:   getVersion(),
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: goVersion,
	}
}

func (i Info) String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)", i.Version, i.GitCommit, i.BuildDate, i.GoVersion)
}

func getVersion() string {
	if Version != devVersion {
		return Version
	}

	// Try to get version from build info (when using go install)
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) >= shortSHALength {
					return fmt.Sprintf("%s-%s", devVersion, setting.Value[:shortSHALength])
				}

				return fmt.Sprintf("%s-%s", devVersion, setting.Value)
			}
		}
	}

	return Version
}
