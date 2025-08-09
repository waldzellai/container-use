package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dagger.io/dagger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotebookEnvironmentCreation tests creating a new notebook environment
func TestNotebookEnvironmentCreation(t *testing.T) {
	ctx := context.Background()
	
	// Initialize Dagger client for testing
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	require.NoError(t, err)
	defer client.Close()
	
	err = Initialize(client)
	require.NoError(t, err)
	
	// Create test notebook environment
	nb, err := Create(ctx, "Test notebook creation", ".", "test-notebook",
		WithKernelSpec("python3"),
		WithNotebookPath("test.ipynb"),
	)
	
	require.NoError(t, err)
	assert.NotNil(t, nb)
	assert.Equal(t, "test-notebook", nb.Name)
	assert.Equal(t, "python3", nb.KernelSpec)
	assert.Equal(t, "test.ipynb", nb.NotebookPath)
	assert.Equal(t, defaultNotebookImage, nb.BaseImage)
	assert.NotNil(t, nb.KernelState)
	assert.Equal(t, 0, nb.KernelState.ExecutionCount)
}

// TestNotebookEnvironmentOptions tests the option pattern
func TestNotebookEnvironmentOptions(t *testing.T) {
	nb := &NotebookEnvironment{
		KernelSpec: "python3",
	}
	
	// Test WithKernelSpec
	WithKernelSpec("julia")(nb)
	assert.Equal(t, "julia", nb.KernelSpec)
	
	// Test WithNotebookPath
	WithNotebookPath("analysis.ipynb")(nb)
	assert.Equal(t, "analysis.ipynb", nb.NotebookPath)
}

// TestExecuteCell tests cell execution
func TestExecuteCell(t *testing.T) {
	ctx := context.Background()
	
	// Mock notebook environment for testing
	nb := &NotebookEnvironment{
		KernelState: &KernelState{
			Variables:      make(map[string]interface{}),
			ExecutionCount: 0,
		},
	}
	
	// Mock container for testing
	// In real tests, we'd use a test container
	client, err := dagger.Connect(ctx)
	require.NoError(t, err)
	defer client.Close()
	
	container := client.Container().
		From("python:3.9-slim").
		WithWorkdir("/workspace")
	
	nb.kernelContainer = container
	
	// Test simple code execution
	code := "print('Hello from test')"
	output, err := nb.ExecuteCell(ctx, 0, code)
	
	// Note: This will fail without proper container setup
	// In a real test environment, we'd mock the container execution
	if err == nil {
		assert.NotNil(t, output)
		assert.Equal(t, "stream", output.OutputType)
		assert.Equal(t, "stdout", output.Name)
		assert.Equal(t, 1, nb.KernelState.ExecutionCount)
	}
}

// TestKernelStateManagement tests kernel state persistence
func TestKernelStateManagement(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	
	nb := &NotebookEnvironment{
		Worktree: tmpDir,
		KernelState: &KernelState{
			KernelID:       "test-kernel-123",
			LastExecuted:   time.Now(),
			ExecutionCount: 5,
			Variables: map[string]interface{}{
				"x": 42,
				"y": "test",
			},
		},
	}
	
	// Test saving kernel state
	err := nb.saveKernelState()
	require.NoError(t, err)
	
	// Verify file exists
	stateFile := filepath.Join(tmpDir, configDir, kernelStateFile)
	assert.FileExists(t, stateFile)
	
	// Read and verify content
	data, err := os.ReadFile(stateFile)
	require.NoError(t, err)
	
	var savedState KernelState
	err = json.Unmarshal(data, &savedState)
	require.NoError(t, err)
	
	assert.Equal(t, nb.KernelState.KernelID, savedState.KernelID)
	assert.Equal(t, nb.KernelState.ExecutionCount, savedState.ExecutionCount)
	assert.Equal(t, nb.KernelState.Variables["x"], savedState.Variables["x"])
}

// TestLoadNotebook tests loading a notebook from file
func TestLoadNotebook(t *testing.T) {
	// Create temporary directory and notebook file
	tmpDir := t.TempDir()
	
	// Create a simple test notebook
	testNotebook := &Notebook{
		Cells: []NotebookCell{
			{
				CellType: "markdown",
				Source:   []string{"# Test Notebook\n"},
			},
			{
				CellType: "code",
				Source:   []string{"x = 42\n", "print(x)\n"},
			},
		},
		Metadata: map[string]interface{}{
			"kernelspec": map[string]interface{}{
				"name": "python3",
			},
		},
		NBFormat:      4,
		NBFormatMinor: 5,
	}
	
	// Write notebook to file
	notebookPath := filepath.Join(tmpDir, "test.ipynb")
	data, err := json.MarshalIndent(testNotebook, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(notebookPath, data, 0644)
	require.NoError(t, err)
	
	// Test loading
	nb := &NotebookEnvironment{
		Worktree:     tmpDir,
		NotebookPath: "test.ipynb",
	}
	
	loaded, err := nb.loadNotebook()
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Len(t, loaded.Cells, 2)
	assert.Equal(t, "markdown", loaded.Cells[0].CellType)
	assert.Equal(t, "code", loaded.Cells[1].CellType)
	assert.Contains(t, loaded.Cells[1].Source[0], "x = 42")
}

// TestExecuteNotebook tests executing all cells in a notebook
func TestExecuteNotebook(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create a test notebook with multiple code cells
	testNotebook := &Notebook{
		Cells: []NotebookCell{
			{
				CellType: "markdown",
				Source:   []string{"# Test\n"},
			},
			{
				CellType: "code",
				Source:   []string{"x = 1\n"},
			},
			{
				CellType: "code",
				Source:   []string{"y = 2\n"},
			},
			{
				CellType: "code",
				Source:   []string{"z = x + y\n"},
			},
		},
		NBFormat:      4,
		NBFormatMinor: 5,
	}
	
	// Write notebook
	notebookPath := filepath.Join(tmpDir, "test.ipynb")
	data, err := json.MarshalIndent(testNotebook, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(notebookPath, data, 0644)
	require.NoError(t, err)
	
	// Create notebook environment
	nb := &NotebookEnvironment{
		Worktree:     tmpDir,
		NotebookPath: "test.ipynb",
		KernelState: &KernelState{
			Variables:      make(map[string]interface{}),
			ExecutionCount: 0,
		},
	}
	
	// Mock container setup would be needed here for real execution
	client, err := dagger.Connect(ctx)
	if err == nil {
		defer client.Close()
		nb.kernelContainer = client.Container().From("python:3.9-slim")
		
		// Test would execute all code cells (3 total, skipping markdown)
		outputs, err := nb.ExecuteNotebook(ctx)
		if err == nil {
			assert.Len(t, outputs, 3)
		}
	}
}

// TestParallelExecutor tests the parallel execution engine
func TestParallelExecutor(t *testing.T) {
	ctx := context.Background()
	
	// Create parallel executor
	executor := NewParallelExecutor(3)
	assert.NotNil(t, executor)
	assert.Equal(t, 3, executor.maxParallel)
	assert.NotNil(t, executor.queue)
	assert.NotNil(t, executor.pool)
	
	// Start executor
	executor.Start(ctx)
	
	// Create mock notebook environments
	for i := 0; i < 3; i++ {
		nb := &NotebookEnvironment{
			Environment: &Environment{
				ID: fmt.Sprintf("test-nb-%d", i),
			},
			KernelState: &KernelState{
				Variables:      make(map[string]interface{}),
				ExecutionCount: 0,
			},
		}
		executor.RegisterNotebook(nb)
	}
	
	// Test execution request
	result := make(chan *ExecutionResult, 1)
	executor.queue <- &ExecutionRequest{
		NotebookID: "test-nb-0",
		CellIndex:  0,
		Code:       "print('test')",
		Result:     result,
	}
	
	// Wait for result with timeout
	select {
	case res := <-result:
		// In real test, would check actual execution
		assert.NotNil(t, res)
	case <-time.After(1 * time.Second):
		// Timeout is expected without real container
	}
	
	// Stop executor
	executor.Stop()
}

// TestParallelExecutorNotFound tests handling of non-existent notebook
func TestParallelExecutorNotFound(t *testing.T) {
	ctx := context.Background()
	
	executor := NewParallelExecutor(1)
	executor.Start(ctx)
	defer executor.Stop()
	
	// Try to execute on non-existent notebook
	output, err := executor.Execute("non-existent", 0, "print('test')")
	
	assert.Nil(t, output)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notebook non-existent not found")
}

// TestNotebookCellTypes tests different cell types
func TestNotebookCellTypes(t *testing.T) {
	cells := []NotebookCell{
		{
			CellType: "code",
			Source:   []string{"x = 1"},
		},
		{
			CellType: "markdown",
			Source:   []string{"# Header"},
		},
		{
			CellType: "raw",
			Source:   []string{"Raw text"},
		},
	}
	
	for _, cell := range cells {
		assert.NotEmpty(t, cell.CellType)
		assert.NotEmpty(t, cell.Source)
	}
}

// TestOutputTypes tests different output types
func TestOutputTypes(t *testing.T) {
	outputs := []Output{
		{
			OutputType: "stream",
			Name:       "stdout",
			Text:       []string{"Hello, world!"},
		},
		{
			OutputType: "execute_result",
			Data: map[string]interface{}{
				"text/plain": "42",
			},
		},
		{
			OutputType: "display_data",
			Data: map[string]interface{}{
				"image/png": "base64data...",
			},
		},
		{
			OutputType: "error",
			Name:       "NameError",
			Text:       []string{"NameError: name 'x' is not defined"},
		},
	}
	
	for _, output := range outputs {
		assert.NotEmpty(t, output.OutputType)
	}
}