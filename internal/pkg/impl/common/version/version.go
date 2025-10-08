package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const SHORT_SHA_LENGTH = 7
const DEV_VERSION = "dev"
const UNKNOWN_VERSION = "unknown"

// nolint:gochecknoglobals
var (
	Version   = DEV_VERSION
	GitCommit = UNKNOWN_VERSION
	BuildDate = UNKNOWN_VERSION
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
		goVersion = UNKNOWN_VERSION
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
	if Version != DEV_VERSION {
		return Version
	}

	// Try to get version from build info (when using go install)
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) >= SHORT_SHA_LENGTH {
					return fmt.Sprintf("%s-%s", DEV_VERSION, setting.Value[:SHORT_SHA_LENGTH])
				}

				return fmt.Sprintf("%s-%s", DEV_VERSION, setting.Value)
			}
		}
	}

	return Version
}
