package test

import (
	"os"
	"path"
	"runtime"
)

func ChWorkingDirectory() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func GetTestDataFilePath(filename string) string {
	_, basePath, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(basePath), "data", filename)
}
