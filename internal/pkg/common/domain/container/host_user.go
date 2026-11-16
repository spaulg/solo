package container

import "strings"

// Environment variables passed into each service container carrying the
// identity of the host user running solo. The entrypoint uses these to hand
// ownership of files it creates on bind mounts back to the host user, so
// they can be removed from the host without elevated privileges.
const (
	HostUIDEnvVar = "SOLO_HOST_UID"
	HostGIDEnvVar = "SOLO_HOST_GID"
)

// isSoloEnvVar reports whether key is an environment variable internal to
// solo that must not be exposed to the commands run inside the container.
func isSoloEnvVar(key string) bool {
	switch key {
	case HostUIDEnvVar, HostGIDEnvVar:
		return true
	default:
		return false
	}
}

// MaskSoloEnvironment returns a copy of environ (in os.Environ() "KEY=value"
// form) with solo's internal environment variables removed. The input slice
// is not modified.
func MaskSoloEnvironment(environ []string) []string {
	masked := make([]string, 0, len(environ))

	for _, entry := range environ {
		key, _, _ := strings.Cut(entry, "=")
		if isSoloEnvVar(key) {
			continue
		}

		masked = append(masked, entry)
	}

	return masked
}
