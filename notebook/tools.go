package notebook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NotebookTools provides MCP tools for notebook operations
var NotebookTools = []*Tool{
	NotebookCreateTool,
	NotebookExecuteCellTool,
	NotebookExecuteAllTool,
	NotebookGetStateTool,
	NotebookParallelRunTool,
}

// Tool represents an MCP tool
type Tool struct {
	Definition mcp.Tool
	Handler    server.ToolHandlerFunc
}

// NotebookCreateTool creates a new notebook environment
var NotebookCreateTool = &Tool{
	Definition: mcp.Tool{
		Name:        "notebook_create",
		Description: "Create a new notebook execution environment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the notebook environment",
				},
				"notebook_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the notebook file",
				},
				"kernel_spec": map[string]interface{}{
					"type":        "string",
					"description": "Kernel specification (e.g., python3, ir, julia)",
					"default":     "python3",
				},
				"explanation": map[string]interface{}{
					"type":        "string",
					"description": "Explanation of what this notebook will do",
				},
			},
			Required: []string{"name", "explanation"},
		},
	},
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			Name         string `json:"name"`
			NotebookPath string `json:"notebook_path"`
			KernelSpec   string `json:"kernel_spec"`
			Explanation  string `json:"explanation"`
		}
		
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return mcp.NewToolResultError("Invalid parameters", err.Error()), nil
		}
		
		// Create notebook environment
		opts := []Option{}
		if params.NotebookPath != "" {
			opts = append(opts, WithNotebookPath(params.NotebookPath))
		}
		if params.KernelSpec != "" {
			opts = append(opts, WithKernelSpec(params.KernelSpec))
		}
		
		nb, err := Create(ctx, params.Explanation, ".", params.Name, opts...)
		if err != nil {
			return mcp.NewToolResultError("Failed to create notebook", err.Error()), nil
		}
		
		// Return notebook info
		result := map[string]interface{}{
			"id":           nb.ID,
			"name":         nb.Name,
			"kernel_spec":  nb.KernelSpec,
			"notebook_path": nb.NotebookPath,
			"workdir":      nb.Workdir,
		}
		
		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	},
}

// NotebookExecuteCellTool executes a specific cell in a notebook
var NotebookExecuteCellTool = &Tool{
	Definition: mcp.Tool{
		Name:        "notebook_execute_cell",
		Description: "Execute a specific cell in a notebook environment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"notebook_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the notebook environment",
				},
				"cell_index": map[string]interface{}{
					"type":        "integer",
					"description": "Index of the cell to execute",
				},
				"code": map[string]interface{}{
					"type":        "string",
					"description": "Code to execute in the cell",
				},
			},
			Required: []string{"notebook_id", "cell_index", "code"},
		},
	},
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			NotebookID string `json:"notebook_id"`
			CellIndex  int    `json:"cell_index"`
			Code       string `json:"code"`
		}
		
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return mcp.NewToolResultError("Invalid parameters", err.Error()), nil
		}
		
		// Get notebook from pool (simplified for prototype)
		// In full implementation, this would look up from a registry
		nb := &NotebookEnvironment{} // Placeholder
		
		output, err := nb.ExecuteCell(ctx, params.CellIndex, params.Code)
		if err != nil {
			return mcp.NewToolResultError("Failed to execute cell", err.Error()), nil
		}
		
		// Format output
		result := map[string]interface{}{
			"cell_index": params.CellIndex,
			"output":     output,
		}
		
		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	},
}

// NotebookExecuteAllTool executes all cells in a notebook
var NotebookExecuteAllTool = &Tool{
	Definition: mcp.Tool{
		Name:        "notebook_execute_all",
		Description: "Execute all cells in a notebook environment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"notebook_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the notebook environment",
				},
			},
			Required: []string{"notebook_id"},
		},
	},
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			NotebookID string `json:"notebook_id"`
		}
		
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return mcp.NewToolResultError("Invalid parameters", err.Error()), nil
		}
		
		// Get notebook from pool
		nb := &NotebookEnvironment{} // Placeholder
		
		outputs, err := nb.ExecuteNotebook(ctx)
		if err != nil {
			return mcp.NewToolResultError("Failed to execute notebook", err.Error()), nil
		}
		
		// Format outputs
		result := map[string]interface{}{
			"notebook_id": params.NotebookID,
			"outputs":     outputs,
			"cell_count":  len(outputs),
		}
		
		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	},
}

// NotebookGetStateTool gets the current state of a notebook kernel
var NotebookGetStateTool = &Tool{
	Definition: mcp.Tool{
		Name:        "notebook_get_state",
		Description: "Get the current state of a notebook kernel",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"notebook_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the notebook environment",
				},
			},
			Required: []string{"notebook_id"},
		},
	},
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			NotebookID string `json:"notebook_id"`
		}
		
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return mcp.NewToolResultError("Invalid parameters", err.Error()), nil
		}
		
		// Get notebook from pool
		nb := &NotebookEnvironment{} // Placeholder
		
		state := nb.GetState()
		
		data, _ := json.Marshal(state)
		return mcp.NewToolResultText(string(data)), nil
	},
}

// NotebookParallelRunTool runs multiple notebooks in parallel
var NotebookParallelRunTool = &Tool{
	Definition: mcp.Tool{
		Name:        "notebook_parallel_run",
		Description: "Run multiple notebook environments in parallel",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"notebooks": map[string]interface{}{
					"type":        "array",
					"description": "List of notebook configurations to run",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Name of the notebook",
							},
							"path": map[string]interface{}{
								"type":        "string",
								"description": "Path to the notebook file",
							},
							"kernel_spec": map[string]interface{}{
								"type":        "string",
								"description": "Kernel specification",
							},
						},
						"required": []string{"name", "path"},
					},
				},
				"max_parallel": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of notebooks to run in parallel",
					"default":     5,
				},
			},
			Required: []string{"notebooks"},
		},
	},
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			Notebooks []struct {
				Name       string `json:"name"`
				Path       string `json:"path"`
				KernelSpec string `json:"kernel_spec"`
			} `json:"notebooks"`
			MaxParallel int `json:"max_parallel"`
		}
		
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return mcp.NewToolResultError("Invalid parameters", err.Error()), nil
		}
		
		if params.MaxParallel == 0 {
			params.MaxParallel = 5
		}
		
		// Create parallel executor
		executor := NewParallelExecutor(params.MaxParallel)
		executor.Start(ctx)
		defer executor.Stop()
		
		// Create and register notebooks
		results := []map[string]interface{}{}
		for _, nbConfig := range params.Notebooks {
			// Create notebook
			opts := []Option{WithNotebookPath(nbConfig.Path)}
			if nbConfig.KernelSpec != "" {
				opts = append(opts, WithKernelSpec(nbConfig.KernelSpec))
			}
			
			nb, err := Create(ctx, fmt.Sprintf("Parallel execution of %s", nbConfig.Name), 
				".", nbConfig.Name, opts...)
			if err != nil {
				results = append(results, map[string]interface{}{
					"name":  nbConfig.Name,
					"error": err.Error(),
				})
				continue
			}
			
			executor.RegisterNotebook(nb)
			
			// Execute notebook
			outputs, err := nb.ExecuteNotebook(ctx)
			if err != nil {
				results = append(results, map[string]interface{}{
					"name":  nbConfig.Name,
					"id":    nb.ID,
					"error": err.Error(),
				})
			} else {
				results = append(results, map[string]interface{}{
					"name":        nbConfig.Name,
					"id":          nb.ID,
					"cell_count":  len(outputs),
					"status":      "completed",
				})
			}
		}
		
		// Return results
		data, _ := json.Marshal(map[string]interface{}{
			"notebooks_run": len(params.Notebooks),
			"results":       results,
		})
		return mcp.NewToolResultText(string(data)), nil
	},
}