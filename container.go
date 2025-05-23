package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"dagger.io/dagger"
	"github.com/google/uuid"
)

type Container struct {
	ID      string
	Workdir string

	mu    sync.Mutex
	state *dagger.Container
}

var containers = map[string]*Container{}

func LoadContainers() error {
	ctr, err := loadState()
	if err != nil {
		return err
	}
	containers = ctr
	return nil
}

func CreateContainer(image string, workdir string) *Container {
	id := uuid.New().String()
	container := &Container{
		ID:      id,
		Workdir: workdir,

		state: dag.Container().
			From(image).
			WithMountedDirectory(workdir, dag.Host().Directory(workdir)).
			WithWorkdir(workdir),
	}
	containers[container.ID] = container
	if err := saveState(container); err != nil {
		panic(err)
	}
	return container
}

func GetContainer(id string) *Container {
	return containers[id]
}

func ListContainers() []*Container {
	ctr := make([]*Container, 0, len(containers))
	for _, container := range containers {
		ctr = append(ctr, container)
	}
	return ctr
}

func (s *Container) RunCmd(ctx context.Context, command string, shell string) (string, error) {
	newState := s.state.WithExec([]string{shell, "-c", command})
	stdout, err := newState.Stdout(ctx)
	if err != nil {
		var exitErr *dagger.ExecError
		if errors.As(err, &exitErr) {
			return fmt.Sprintf("command failed with exit code %d.\nstdout: %s\nstderr: %s", exitErr.ExitCode, exitErr.Stdout, exitErr.Stderr), nil
		}
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = newState
	if err := saveState(s); err != nil {
		return "", err
	}
	return stdout, nil
}

func (s *Container) ReadFile(ctx context.Context, targetFile string, shouldReadEntireFile bool, startLineOneIndexed int, endLineOneIndexedInclusive int) (string, error) {
	file, err := s.state.File(targetFile).Contents(ctx)
	if err != nil {
		return "", err
	}
	if shouldReadEntireFile {
		return string(file), err
	}

	lines := strings.Split(string(file), "\n")
	start := startLineOneIndexed - 1
	end := endLineOneIndexedInclusive
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n"), nil
}
