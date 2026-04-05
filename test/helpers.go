package test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
)

func GetModuleName() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errors.New("failed to detect module name")
	}

	return info.Main.Path, nil
}

func GetProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to detect project root")
	}

	return filepath.Join(filepath.Dir(filename), ".."), nil
}

func ChWorkingDirectory() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func GetTestDataFilePath(filename string) string {
	_, basePath, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(basePath), "data", filename)
}
