package context

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strconv"
	"time"

	commoncontainer "github.com/spaulg/solo/internal/pkg/common/domain/container"
	"github.com/spaulg/solo/internal/pkg/common/infra/logging"
)

const initialHostnameFile = "/solo/container/data/initial_hostname"
const logFilePath = "/solo/service/logs"

type EntrypointContext struct {
	Logger          *slog.Logger
	InitialHostname string
}

func LoadEntrypointContext() (*EntrypointContext, error) {
	initialHostname, err := readInitialHostname()
	if err != nil {
		return nil, err
	}

	logFileName := path.Join(logFilePath, time.Now().Format("2006-01-02"), initialHostname+".log")

	if err := os.MkdirAll(path.Dir(logFileName), 0777); err != nil {
		return nil, err
	}

	builder := logging.NewLogHandlerBuilder()
	handler, err := builder.
		WithLogFilePath(logFileName).
		WithLogLevel("info").
		WithLogHandlerName("text").
		Build()

	if err != nil {
		panic(fmt.Sprintf("%v\n", err))
	}

	logger := slog.New(handler)

	// Files written to the log bind mount are created as root.
	// Hand them back to the host user so solo can remove them.
	if err := chownToHostUser(path.Dir(logFileName), logFileName); err != nil {
		logger.Warn(fmt.Sprintf("Failed to change log file ownership to host user: %v", err))
	}

	return &EntrypointContext{
		Logger:          logger,
		InitialHostname: initialHostname,
	}, nil
}

// chownToHostUser changes ownership of the given paths to the host user
// identified by the SOLO_HOST_UID/SOLO_HOST_GID environment variables.
// It is a no-op when not running as root or when the variables are absent.
func chownToHostUser(paths ...string) error {
	if os.Geteuid() != 0 {
		return nil
	}

	uidStr, uidOk := os.LookupEnv(commoncontainer.HostUIDEnvVar)
	gidStr, gidOk := os.LookupEnv(commoncontainer.HostGIDEnvVar)
	if !uidOk || !gidOk {
		return nil
	}

	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", commoncontainer.HostUIDEnvVar, uidStr, err)
	}

	gid, err := strconv.Atoi(gidStr)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", commoncontainer.HostGIDEnvVar, gidStr, err)
	}

	for _, p := range paths {
		if err := os.Chown(p, uid, gid); err != nil {
			return err
		}
	}

	return nil
}

func readInitialHostname() (string, error) {
	_, err := os.Stat(initialHostnameFile)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}

		if err := os.MkdirAll(path.Dir(initialHostnameFile), 0777); err != nil {
			return "", err
		}

		hostname, err := os.Hostname()
		if err != nil {
			return "", err
		}

		if err := os.WriteFile(initialHostnameFile, []byte(hostname), 0600); err != nil {
			return "", err
		}

		return hostname, nil
	}

	hostnameBytes, err := os.ReadFile(initialHostnameFile)
	if err != nil {
		return "", err
	}

	return string(hostnameBytes), nil
}
