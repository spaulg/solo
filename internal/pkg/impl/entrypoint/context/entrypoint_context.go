package context

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/spaulg/solo/internal/pkg/impl/common/logging"
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

	return &EntrypointContext{
		Logger:          slog.New(handler),
		InitialHostname: initialHostname,
	}, nil
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
