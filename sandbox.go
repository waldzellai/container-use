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

type Sandbox struct {
	ID      string
	Workdir string

	mu    sync.Mutex
	state *dagger.Container
}

var sandboxes = map[string]*Sandbox{}

func CreateSandbox(image string, workdir string) *Sandbox {
	id := uuid.New().String()
	sandbox := &Sandbox{
		ID:      id,
		Workdir: workdir,

		state: dag.Container().
			From(image).
			WithMountedDirectory(workdir, dag.Host().Directory(workdir)).
			WithWorkdir(workdir),
	}
	sandboxes[sandbox.ID] = sandbox
	return sandbox
}

func GetSandbox(id string) *Sandbox {
	return sandboxes[id]
}

func ListSandboxes() []*Sandbox {
	sandboxes := make([]*Sandbox, 0, len(sandboxes))
	for _, sandbox := range sandboxes {
		sandboxes = append(sandboxes, sandbox)
	}
	return sandboxes
}

func (s *Sandbox) RunTerminalCmd(ctx context.Context, command string) (string, error) {
	newState := s.state.WithExec([]string{"/bin/bash", "-c", command})
	stdout, err := newState.Stdout(ctx)
	if err != nil {
		var exitErr *dagger.ExecError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("command failed with exit code %d.\nstdout: %s\nstderr: %s", exitErr.ExitCode, exitErr.Stdout, exitErr.Stderr)
		}
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = newState
	return stdout, nil
}

func (s *Sandbox) ReadFile(ctx context.Context, targetFile string, shouldReadEntireFile bool, startLineOneIndexed int, endLineOneIndexedInclusive int) (string, error) {
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
