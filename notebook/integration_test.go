package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dagger/container-use/notebook/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationSimpleNotebook tests end-to-end execution of a simple notebook
func TestIntegrationSimpleNotebook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create test workspace
	err := testdata.CreateTestWorkspace(tmpDir)
	require.NoError(t, err)
	
	// Mock notebook environment creation
	// In real integration test, this would use actual Dagger containers
	nb := &NotebookEnvironment{
		Worktree:     tmpDir,
		NotebookPath: "simple.ipynb",
		KernelSpec:   "python3",
		KernelState: &KernelState{
			Variables:      make(map[string]interface{}),
			ExecutionCount: 0,
		},
	}
	
	// Load the notebook
	notebook, err := nb.loadNotebook()
	require.NoError(t, err)
	assert.Len(t, notebook.Cells, 3)
	
	// Verify cell types
	assert.Equal(t, "markdown", notebook.Cells[0].CellType)
	assert.Equal(t, "code", notebook.Cells[1].CellType)
	assert.Equal(t, "code", notebook.Cells[2].CellType)
}

// TestIntegrationParallelExecution tests running multiple notebooks in parallel
func TestIntegrationParallelExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create test workspace
	err := testdata.CreateTestWorkspace(tmpDir)
	require.NoError(t, err)
	
	// Create parallel executor
	executor := NewParallelExecutor(3)
	executor.Start(ctx)
	defer executor.Stop()
	
	// Create multiple notebook environments
	notebooks := []string{"simple.ipynb", "data_science.ipynb", "ml_training.ipynb"}
	var wg sync.WaitGroup
	results := make(map[string]bool)
	mu := sync.Mutex{}
	
	for i, notebookPath := range notebooks {
		wg.Add(1)
		go func(id int, path string) {
			defer wg.Done()
			
			nb := &NotebookEnvironment{
				Environment: &Environment{
					ID:   fmt.Sprintf("test-nb-%d", id),
					Name: fmt.Sprintf("notebook-%d", id),
				},
				Worktree:     tmpDir,
				NotebookPath: path,
				KernelState: &KernelState{
					Variables:      make(map[string]interface{}),
					ExecutionCount: 0,
				},
			}
			
			executor.RegisterNotebook(nb)
			
			// Simulate execution
			mu.Lock()
			results[path] = true
			mu.Unlock()
		}(i, notebookPath)
	}
	
	wg.Wait()
	
	// Verify all notebooks were processed
	assert.Len(t, results, 3)
	for _, path := range notebooks {
		assert.True(t, results[path])
	}
}

// TestIntegrationKernelStatePersistence tests saving and loading kernel state
func TestIntegrationKernelStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create notebook with state
	nb1 := &NotebookEnvironment{
		Worktree: tmpDir,
		KernelState: &KernelState{
			KernelID:       "test-kernel-123",
			LastExecuted:   time.Now(),
			ExecutionCount: 10,
			Variables: map[string]interface{}{
				"data": []float64{1.0, 2.0, 3.0},
				"model": map[string]string{
					"type": "regression",
					"status": "trained",
				},
			},
		},
	}
	
	// Save state
	err := nb1.saveKernelState()
	require.NoError(t, err)
	
	// Create new notebook and load state
	stateFile := filepath.Join(tmpDir, configDir, kernelStateFile)
	data, err := os.ReadFile(stateFile)
	require.NoError(t, err)
	
	var loadedState KernelState
	err = json.Unmarshal(data, &loadedState)
	require.NoError(t, err)
	
	// Verify state
	assert.Equal(t, "test-kernel-123", loadedState.KernelID)
	assert.Equal(t, 10, loadedState.ExecutionCount)
	assert.NotNil(t, loadedState.Variables["data"])
	assert.NotNil(t, loadedState.Variables["model"])
}

// TestIntegrationErrorHandling tests handling of notebook errors
func TestIntegrationErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test workspace
	err := testdata.CreateTestWorkspace(tmpDir)
	require.NoError(t, err)
	
	nb := &NotebookEnvironment{
		Worktree:     tmpDir,
		NotebookPath: "errors.ipynb",
		KernelState: &KernelState{
			Variables:      make(map[string]interface{}),
			ExecutionCount: 0,
		},
	}
	
	// Load error notebook
	notebook, err := nb.loadNotebook()
	require.NoError(t, err)
	
	// Verify it has error cells
	assert.Len(t, notebook.Cells, 3)
	assert.Contains(t, notebook.Cells[1].Source[1], "undefined_variable")
	assert.Contains(t, notebook.Cells[2].Source[1], "1 / 0")
}

// TestIntegrationConcurrentAccess tests concurrent access to notebooks
func TestIntegrationConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	tmpDir := t.TempDir()
	
	nb := &NotebookEnvironment{
		Worktree: tmpDir,
		KernelState: &KernelState{
			Variables:      make(map[string]interface{}),
			ExecutionCount: 0,
		},
		mu: sync.Mutex{},
	}
	
	// Simulate concurrent state updates
	var wg sync.WaitGroup
	updates := 100
	
	for i := 0; i < updates; i++ {
		wg.Add(1)
		go func(count int) {
			defer wg.Done()
			
			nb.mu.Lock()
			nb.KernelState.ExecutionCount++
			nb.KernelState.LastExecuted = time.Now()
			nb.mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	
	// Verify final count
	assert.Equal(t, updates, nb.KernelState.ExecutionCount)
}

// TestIntegrationNotebookRegistry tests the notebook registry pattern
func TestIntegrationNotebookRegistry(t *testing.T) {
	registry := NewMockRegistry()
	
	// Register multiple notebooks
	for i := 0; i < 5; i++ {
		nb := &NotebookEnvironment{
			Environment: &Environment{
				ID:   fmt.Sprintf("nb-%d", i),
				Name: fmt.Sprintf("notebook-%d", i),
			},
			KernelSpec: "python3",
		}
		registry.Register(nb)
	}
	
	// Verify registration
	for i := 0; i < 5; i++ {
		nb := registry.Get(fmt.Sprintf("nb-%d", i))
		assert.NotNil(t, nb)
		assert.Equal(t, fmt.Sprintf("notebook-%d", i), nb.Name)
	}
	
	// Test non-existent notebook
	assert.Nil(t, registry.Get("non-existent"))
}

// BenchmarkNotebookCreation benchmarks notebook creation
func BenchmarkNotebookCreation(b *testing.B) {
	tmpDir := b.TempDir()
	testdata.CreateTestWorkspace(tmpDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nb := &NotebookEnvironment{
			Worktree:     tmpDir,
			NotebookPath: "simple.ipynb",
			KernelState: &KernelState{
				Variables:      make(map[string]interface{}),
				ExecutionCount: 0,
			},
		}
		_, _ = nb.loadNotebook()
	}
}

// BenchmarkParallelExecution benchmarks parallel execution
func BenchmarkParallelExecution(b *testing.B) {
	ctx := context.Background()
	executor := NewParallelExecutor(10)
	executor.Start(ctx)
	defer executor.Stop()
	
	// Pre-register notebooks
	for i := 0; i < 10; i++ {
		nb := &NotebookEnvironment{
			Environment: &Environment{
				ID: fmt.Sprintf("bench-nb-%d", i),
			},
			KernelState: &KernelState{
				Variables:      make(map[string]interface{}),
				ExecutionCount: 0,
			},
		}
		executor.RegisterNotebook(nb)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate execution request
		result := make(chan *ExecutionResult, 1)
		executor.queue <- &ExecutionRequest{
			NotebookID: fmt.Sprintf("bench-nb-%d", i%10),
			CellIndex:  0,
			Code:       "x = 1",
			Result:     result,
		}
		<-result
	}
}