# Notebook-Use: Design Document for Parallel Notebook Execution

## Executive Summary

This document outlines the transformation of container-use to notebook-use, enabling notebooks (Jupyter/IPython) to become first-class entities that can run many headless instances in parallel, similar to how container-use manages containerized environments.

## Current Container-Use Architecture

### Key Components

1. **Environment Management** (`environment/environment.go`)
   - Manages container lifecycle with Dagger
   - Tracks container state and history
   - Handles git-based persistence
   - Supports parallel container execution

2. **Git Integration** (`environment/git.go`)
   - Uses git worktrees for isolation
   - Pushes changes to `container-use/` remote branches
   - Maintains history through git notes

3. **MCP Server** (`mcpserver/tools.go`)
   - Provides tools for container interaction
   - Handles file operations, command execution
   - Manages environment lifecycle

4. **Key Features**
   - Parallel execution of multiple containers
   - State persistence and versioning
   - Git-based change tracking
   - MCP-based tool interface

## Proposed Notebook-Use Architecture

### Core Concept

Transform notebooks from static documents to executable, stateful environments that can:
- Run multiple instances in parallel
- Maintain kernel state across operations
- Track changes and outputs
- Integrate with git for versioning

### Architecture Components

#### 1. Notebook Environment (`notebook/environment.go`)

```go
type NotebookEnvironment struct {
    ID           string
    Name         string
    Source       string      // Source notebook file
    Worktree     string      // Git worktree path
    
    // Notebook-specific fields
    KernelSpec   string      // e.g., "python3", "ir", "julia"
    Runtime      string      // e.g., "jupyter", "ipykernel"
    BaseImage    string      // Container image with Jupyter
    
    // State management
    History      History
    KernelState  *KernelState
    
    // Container backend
    container    *dagger.Container
    kernelID     string      // Running kernel ID
}

type KernelState struct {
    Variables    map[string]interface{}
    Imports      []string
    CellOutputs  map[int]CellOutput
}
```

#### 2. Kernel Management (`notebook/kernel.go`)

- Manage Jupyter kernel lifecycle
- Handle kernel communication via ZMQ
- Execute cells and capture outputs
- Maintain kernel state between executions

#### 3. Parallel Execution Engine

```go
type NotebookPool struct {
    environments map[string]*NotebookEnvironment
    maxParallel  int
    queue        chan *ExecutionRequest
}

type ExecutionRequest struct {
    NotebookID   string
    CellIndex    int
    Code         string
    Callback     func(output CellOutput, err error)
}
```

#### 4. MCP Tools for Notebooks

New tools to add:
- `notebook_create` - Create new notebook environment
- `notebook_execute_cell` - Execute specific cell
- `notebook_execute_all` - Execute entire notebook
- `notebook_get_state` - Get current kernel state
- `notebook_parallel_run` - Run multiple notebooks in parallel
- `notebook_export` - Export results (HTML, PDF, etc.)

### Implementation Strategy

#### Phase 1: Core Infrastructure

1. **Notebook Environment Manager**
   - Adapt `environment.go` for notebooks
   - Implement notebook-specific state tracking
   - Create kernel lifecycle management

2. **Jupyter Integration**
   - Use Jupyter Server API for headless execution
   - Implement kernel protocol client
   - Handle cell execution and output capture

3. **Container Setup**
   - Create base images with Jupyter + common kernels
   - Support custom environments via requirements.txt/environment.yml
   - Enable GPU support for ML workloads

#### Phase 2: Parallel Execution

1. **Execution Queue**
   - Implement work queue for parallel notebook execution
   - Resource management (CPU, memory limits)
   - Execution prioritization

2. **State Management**
   - Checkpoint kernel state between executions
   - Enable state sharing between notebook instances
   - Implement rollback/recovery mechanisms

3. **Output Handling**
   - Stream outputs in real-time
   - Aggregate results from parallel runs
   - Support different output formats

#### Phase 3: Advanced Features

1. **Distributed Execution**
   - Support multi-node execution
   - Implement data parallelism for large datasets
   - Enable model parallel training

2. **Notebook Composition**
   - Chain notebooks together
   - Share state between notebooks
   - Create notebook pipelines

3. **Integration Features**
   - Version control integration
   - CI/CD pipeline support
   - Monitoring and observability

### Key Transformations from Container-Use

1. **From Shell Commands to Cell Execution**
   ```go
   // Container-use
   env.Run(ctx, "python script.py", explanation)
   
   // Notebook-use
   notebook.ExecuteCell(ctx, cellIndex, explanation)
   ```

2. **From File-Based to Cell-Based Operations**
   - Cells become the primary unit of work
   - Output tracking per cell
   - State persistence at cell granularity

3. **From Process-Based to Kernel-Based**
   - Long-running kernels maintain state
   - Interactive execution model
   - Rich output support (plots, HTML, etc.)

### Example Usage

```go
// Create multiple notebook environments
notebooks := []string{"model-training", "data-analysis", "visualization"}
envs := make([]*NotebookEnvironment, len(notebooks))

for i, name := range notebooks {
    env, err := notebook.Create(ctx, name, "experiments.ipynb")
    if err != nil {
        return err
    }
    envs[i] = env
}

// Execute in parallel
results := make(chan Result, len(envs))
for _, env := range envs {
    go func(nb *NotebookEnvironment) {
        output, err := nb.ExecuteAll(ctx)
        results <- Result{nb.Name, output, err}
    }(env)
}

// Collect results
for i := 0; i < len(envs); i++ {
    result := <-results
    fmt.Printf("Notebook %s completed\n", result.Name)
}
```

### Benefits

1. **Scalability**: Run hundreds of notebook instances in parallel
2. **Reproducibility**: Git-tracked changes and environment state
3. **Flexibility**: Mix different kernels and environments
4. **Integration**: MCP-based tools for AI assistants
5. **Efficiency**: Reuse kernel state across executions

### Challenges and Solutions

1. **Kernel State Management**
   - Challenge: Kernels are stateful and memory-intensive
   - Solution: Implement kernel pooling and state checkpointing

2. **Output Size**
   - Challenge: Notebooks can generate large outputs
   - Solution: Stream outputs, implement pagination, store in object storage

3. **Resource Management**
   - Challenge: Multiple kernels consume significant resources
   - Solution: Resource quotas, automatic scaling, kernel recycling

4. **Synchronization**
   - Challenge: Coordinating parallel notebook execution
   - Solution: Event-driven architecture, message queuing

### Migration Path

1. **Compatibility Layer**
   - Maintain container-use API compatibility
   - Gradual migration of tools
   - Dual-mode operation

2. **Tool Migration**
   ```go
   // Adapter pattern for existing tools
   func (n *NotebookEnvironment) Run(cmd string) (string, error) {
       // Convert shell command to notebook cell
       return n.ExecuteCell(ctx, -1, fmt.Sprintf("!%s", cmd))
   }
   ```

3. **Testing Strategy**
   - Unit tests for kernel management
   - Integration tests for parallel execution
   - Performance benchmarks

### Next Steps

1. **Prototype Implementation**
   - Basic notebook environment
   - Single kernel execution
   - MCP tool integration

2. **Validation**
   - Performance testing
   - Use case validation
   - User feedback

3. **Full Implementation**
   - Complete feature set
   - Production hardening
   - Documentation

## Conclusion

Transforming container-use to notebook-use will enable powerful parallel notebook execution capabilities while maintaining the core benefits of the container-use architecture. This design provides a clear path forward with incremental implementation phases and addresses key technical challenges.