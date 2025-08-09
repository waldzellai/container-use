package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"dagger.io/dagger"
)

var dag *dagger.Client

const (
	defaultNotebookImage = "jupyter/base-notebook:latest"
	configDir            = ".notebook-use"
	notebookFile         = "notebook.ipynb"
	kernelStateFile      = "kernel-state.json"
)

// Initialize sets up the Dagger client
func Initialize(client *dagger.Client) error {
	dag = client
	return nil
}

// NotebookCell represents a single cell in a notebook
type NotebookCell struct {
	CellType   string   `json:"cell_type"`
	Source     []string `json:"source"`
	Outputs    []Output `json:"outputs,omitempty"`
	ExecutionCount *int `json:"execution_count,omitempty"`
}

// Output represents cell output
type Output struct {
	OutputType string                 `json:"output_type"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Text       []string               `json:"text,omitempty"`
	Name       string                 `json:"name,omitempty"`
}

// KernelState tracks the state of a Jupyter kernel
type KernelState struct {
	KernelID       string                 `json:"kernel_id"`
	LastExecuted   time.Time              `json:"last_executed"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	ExecutionCount int                    `json:"execution_count"`
}

// Environment represents a base environment (simplified from container-use)
type Environment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Worktree string `json:"worktree"`
	Workdir  string `json:"workdir"`
	BaseImage string `json:"base_image"`
}

// NotebookEnvironment represents a notebook execution environment
type NotebookEnvironment struct {
	*Environment
	
	// Notebook-specific fields
	NotebookPath string       `json:"notebook_path"`
	KernelSpec   string       `json:"kernel_spec"`
	KernelState  *KernelState `json:"kernel_state,omitempty"`
	
	// Runtime state
	kernelContainer *dagger.Container
	kernelPort      int
	mu              sync.Mutex
}

// Create creates a new notebook environment
func Create(ctx context.Context, explanation, source, name string, opts ...Option) (*NotebookEnvironment, error) {
	// Create base environment (simplified version)
	baseEnv := &Environment{
		ID:        fmt.Sprintf("%s-%d", name, time.Now().Unix()),
		Name:      name,
		Source:    source,
		Worktree:  fmt.Sprintf("/tmp/notebook-use/%s", name),
		Workdir:   "/home/jovyan/work",
		BaseImage: defaultNotebookImage,
	}
	
	nb := &NotebookEnvironment{
		Environment:  baseEnv,
		KernelSpec:   "python3",
		KernelState:  &KernelState{
			Variables:      make(map[string]interface{}),
			ExecutionCount: 0,
		},
	}
	
	// Apply options
	for _, opt := range opts {
		opt(nb)
	}
	
	// Set up notebook-specific configuration
	nb.BaseImage = defaultNotebookImage
	nb.Workdir = "/home/jovyan/work"
	
	// Initialize kernel container
	if err := nb.initializeKernel(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize kernel: %w", err)
	}
	
	return nb, nil
}

// Option configures a NotebookEnvironment
type Option func(*NotebookEnvironment)

// WithKernelSpec sets the kernel specification
func WithKernelSpec(spec string) Option {
	return func(nb *NotebookEnvironment) {
		nb.KernelSpec = spec
	}
}

// WithNotebookPath sets the path to the notebook file
func WithNotebookPath(path string) Option {
	return func(nb *NotebookEnvironment) {
		nb.NotebookPath = path
	}
}

// initializeKernel sets up the Jupyter kernel container
func (nb *NotebookEnvironment) initializeKernel(ctx context.Context) error {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	
	// Build kernel container with Jupyter
	container := dag.
		Container().
		From(nb.BaseImage).
		WithWorkdir(nb.Workdir).
		WithExposedPort(8888). // Jupyter server port
		WithExec([]string{
			"jupyter", "notebook", 
			"--ip=0.0.0.0", 
			"--port=8888", 
			"--no-browser", 
			"--allow-root",
			"--NotebookApp.token=''",
			"--NotebookApp.password=''",
		})
	
	nb.kernelContainer = container
	nb.kernelPort = 8888
	
	return nil
}

// ExecuteCell executes a specific cell in the notebook
func (nb *NotebookEnvironment) ExecuteCell(ctx context.Context, cellIndex int, code string) (*Output, error) {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	
	// For prototype, we'll execute code directly in container
	// In full implementation, this would use Jupyter kernel protocol
	
	slog.Info("Executing notebook cell", "index", cellIndex, "env", nb.ID)
	
	// Create a Python script from the cell code
	scriptPath := fmt.Sprintf("/tmp/cell_%d.py", cellIndex)
	
	// Write the code to a file in the container
	container := nb.kernelContainer.
		WithNewFile(scriptPath, dagger.ContainerWithNewFileOpts{
			Contents: code,
		})
	
	// Execute the script and capture output
	result, err := container.
		WithExec([]string{"python", scriptPath}).
		Stdout(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to execute cell: %w", err)
	}
	
	// Update execution count
	nb.KernelState.ExecutionCount++
	nb.KernelState.LastExecuted = time.Now()
	
	// Create output object
	output := &Output{
		OutputType: "stream",
		Name:       "stdout",
		Text:       []string{result},
	}
	
	// Save state
	if err := nb.saveKernelState(); err != nil {
		slog.Warn("Failed to save kernel state", "error", err)
	}
	
	return output, nil
}

// ExecuteNotebook executes all cells in the notebook
func (nb *NotebookEnvironment) ExecuteNotebook(ctx context.Context) ([]*Output, error) {
	// Load notebook
	notebook, err := nb.loadNotebook()
	if err != nil {
		return nil, fmt.Errorf("failed to load notebook: %w", err)
	}
	
	outputs := make([]*Output, 0, len(notebook.Cells))
	
	// Execute each code cell
	for i, cell := range notebook.Cells {
		if cell.CellType != "code" {
			continue
		}
		
		// Join source lines
		code := ""
		for _, line := range cell.Source {
			code += line
		}
		
		output, err := nb.ExecuteCell(ctx, i, code)
		if err != nil {
			return outputs, fmt.Errorf("failed to execute cell %d: %w", i, err)
		}
		
		outputs = append(outputs, output)
	}
	
	return outputs, nil
}

// GetState returns the current kernel state
func (nb *NotebookEnvironment) GetState() *KernelState {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	return nb.KernelState
}

// loadNotebook loads a notebook from file
func (nb *NotebookEnvironment) loadNotebook() (*Notebook, error) {
	notebookPath := filepath.Join(nb.Worktree, nb.NotebookPath)
	
	data, err := os.ReadFile(notebookPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %w", err)
	}
	
	var notebook Notebook
	if err := json.Unmarshal(data, &notebook); err != nil {
		return nil, fmt.Errorf("failed to parse notebook: %w", err)
	}
	
	return &notebook, nil
}

// saveKernelState persists the kernel state
func (nb *NotebookEnvironment) saveKernelState() error {
	cfg := path.Join(nb.Worktree, configDir)
	if err := os.MkdirAll(cfg, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(nb.KernelState, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path.Join(cfg, kernelStateFile), data, 0644)
}

// Notebook represents a Jupyter notebook structure
type Notebook struct {
	Cells    []NotebookCell         `json:"cells"`
	Metadata map[string]interface{} `json:"metadata"`
	NBFormat int                    `json:"nbformat"`
	NBFormatMinor int               `json:"nbformat_minor"`
}

// ParallelExecutor manages parallel notebook execution
type ParallelExecutor struct {
	pool        map[string]*NotebookEnvironment
	maxParallel int
	queue       chan *ExecutionRequest
	wg          sync.WaitGroup
}

// ExecutionRequest represents a notebook execution request
type ExecutionRequest struct {
	NotebookID string
	CellIndex  int
	Code       string
	Result     chan *ExecutionResult
}

// ExecutionResult contains the result of an execution
type ExecutionResult struct {
	Output *Output
	Error  error
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxParallel int) *ParallelExecutor {
	return &ParallelExecutor{
		pool:        make(map[string]*NotebookEnvironment),
		maxParallel: maxParallel,
		queue:       make(chan *ExecutionRequest, maxParallel*2),
	}
}

// Start starts the parallel executor
func (pe *ParallelExecutor) Start(ctx context.Context) {
	for i := 0; i < pe.maxParallel; i++ {
		pe.wg.Add(1)
		go pe.worker(ctx)
	}
}

// Stop stops the parallel executor
func (pe *ParallelExecutor) Stop() {
	close(pe.queue)
	pe.wg.Wait()
}

// worker processes execution requests
func (pe *ParallelExecutor) worker(ctx context.Context) {
	defer pe.wg.Done()
	
	for req := range pe.queue {
		nb, ok := pe.pool[req.NotebookID]
		if !ok {
			req.Result <- &ExecutionResult{
				Error: fmt.Errorf("notebook %s not found", req.NotebookID),
			}
			continue
		}
		
		output, err := nb.ExecuteCell(ctx, req.CellIndex, req.Code)
		req.Result <- &ExecutionResult{
			Output: output,
			Error:  err,
		}
	}
}

// RegisterNotebook adds a notebook to the pool
func (pe *ParallelExecutor) RegisterNotebook(nb *NotebookEnvironment) {
	pe.pool[nb.ID] = nb
}

// Execute submits an execution request
func (pe *ParallelExecutor) Execute(notebookID string, cellIndex int, code string) (*Output, error) {
	result := make(chan *ExecutionResult, 1)
	pe.queue <- &ExecutionRequest{
		NotebookID: notebookID,
		CellIndex:  cellIndex,
		Code:       code,
		Result:     result,
	}
	
	res := <-result
	return res.Output, res.Error
}