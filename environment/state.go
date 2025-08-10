package environment

import (
	"encoding/json"
	"fmt"
	"time"
)

type State struct {
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	Config    *EnvironmentConfig `json:"config,omitempty"`
	Container string             `json:"container,omitempty"`
	Title     string             `json:"title,omitempty"`

	BackgroundProcesses []BackgroundProcess `json:"background_processes,omitempty"`
}

// BackgroundProcess records a host-mode background subprocess
type BackgroundProcess struct {
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	Shell     string    `json:"shell"`
	Ports     []int     `json:"ports,omitempty"`
	Workdir   string    `json:"workdir"`
	StartedAt time.Time `json:"started_at"`
}

func (s *State) Marshal() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func (s *State) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, &s); err != nil {
		// Try to migrate the legacy state
		legacySt, err := migrateLegacyState(data)
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}
		*s = *legacySt
	}
	return nil
}

func migrateLegacyState(state []byte) (*State, error) {
	var history legacyState
	if err := json.Unmarshal(state, &history); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	latest := history.Latest()
	if latest == nil {
		return nil, fmt.Errorf("no latest revision found")
	}

	return &State{
		Container: latest.State,
		CreatedAt: latest.CreatedAt,
		UpdatedAt: latest.CreatedAt,
	}, nil
}

type legacyState []*legacyRevision

func (h legacyState) Latest() *legacyRevision {
	if len(h) == 0 {
		return nil
	}
	return h[len(h)-1]
}

func (h legacyState) LatestVersion() int {
	latest := h.Latest()
	if latest == nil {
		return 0
	}
	return latest.Version
}

func (h legacyState) Get(version int) *legacyRevision {
	for _, revision := range h {
		if revision.Version == version {
			return revision
		}
	}
	return nil
}

type legacyRevision struct {
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Explanation string    `json:"explanation"`
	Output      string    `json:"output,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	State       string    `json:"state"`
}
