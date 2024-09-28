package main

import (
	"dagger/solo/internal/dagger"
)

type Solo struct{}

func (m *Solo) Build(sourceDirectory *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/source", sourceDirectory).
		WithExec([]string{"apk", "add", "--no-cache", "go"}).
		WithExec([]string{"mkdir", "-p", "/build"}).
		WithEnvVariable("GOOS", "darwin").
		WithWorkdir("/source/cli").
		WithExec([]string{"go", "build", "-o", "/build/solo", "main.go"}).
		WithWorkdir("/source/agent").
		WithExec([]string{"go", "build", "-o", "/build/solo-entrypoint", "main.go"}).
		WithWorkdir("/source").
		WithExec([]string{"cp", "solo.yml", "/build/solo.yml"}).
		Directory("/build")
}
