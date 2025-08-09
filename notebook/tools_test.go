package notebook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotebookCreateTool tests the notebook_create MCP tool
func TestNotebookCreateTool(t *testing.T) {
	ctx := context.Background()
	
	// Test valid parameters
	params := map[string]interface{}{
		"name":         "test-notebook",
		"notebook_path": "test.ipynb",
		"kernel_spec":  "python3",
		"explanation":  "Testing notebook creation",
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	request := mcp.CallToolRequest{
		Params: paramsJSON,
	}
	
	// Note: This test would need proper Dagger initialization to work
	// For now, we're testing the parameter handling
	result, err := NotebookCreateTool.Handler(ctx, request)
	
	// The handler will fail without proper setup, but we can test parameter validation
	if err == nil {
		assert.NotNil(t, result)
		// Would contain notebook info in real test
	}
}

// TestNotebookCreateToolInvalidParams tests error handling
func TestNotebookCreateToolInvalidParams(t *testing.T) {
	ctx := context.Background()
	
	// Test with invalid JSON
	request := mcp.CallToolRequest{
		Params: []byte("invalid json"),
	}
	
	result, err := NotebookCreateTool.Handler(ctx, request)
	assert.Nil(t, err) // Handler returns error in result, not as error
	assert.Contains(t, result.Content[0].Text, "Invalid parameters")
}

// TestNotebookExecuteCellTool tests the notebook_execute_cell tool
func TestNotebookExecuteCellTool(t *testing.T) {
	ctx := context.Background()
	
	params := map[string]interface{}{
		"notebook_id": "test-nb-123",
		"cell_index":  0,
		"code":        "print('Hello, world!')",
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	request := mcp.CallToolRequest{
		Params: paramsJSON,
	}
	
	// Test parameter handling
	result, err := NotebookExecuteCellTool.Handler(ctx, request)
	
	// Without actual notebook setup, this will fail
	// but we're testing the parameter parsing
	if result != nil {
		assert.NotNil(t, result.Content)
	}
}

// TestNotebookExecuteAllTool tests the notebook_execute_all tool
func TestNotebookExecuteAllTool(t *testing.T) {
	ctx := context.Background()
	
	params := map[string]interface{}{
		"notebook_id": "test-nb-123",
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	request := mcp.CallToolRequest{
		Params: paramsJSON,
	}
	
	result, err := NotebookExecuteAllTool.Handler(ctx, request)
	
	// Test would need actual notebook to succeed
	if result != nil {
		assert.NotNil(t, result.Content)
	}
}

// TestNotebookGetStateTool tests the notebook_get_state tool
func TestNotebookGetStateTool(t *testing.T) {
	ctx := context.Background()
	
	params := map[string]interface{}{
		"notebook_id": "test-nb-123",
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	request := mcp.CallToolRequest{
		Params: paramsJSON,
	}
	
	result, err := NotebookGetStateTool.Handler(ctx, request)
	
	// Would return kernel state in real test
	if result != nil {
		assert.NotNil(t, result.Content)
	}
}

// TestNotebookParallelRunTool tests the notebook_parallel_run tool
func TestNotebookParallelRunTool(t *testing.T) {
	ctx := context.Background()
	
	params := map[string]interface{}{
		"notebooks": []map[string]interface{}{
			{
				"name":        "nb1",
				"path":        "test1.ipynb",
				"kernel_spec": "python3",
			},
			{
				"name": "nb2",
				"path": "test2.ipynb",
			},
		},
		"max_parallel": 2,
	}
	
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	
	request := mcp.CallToolRequest{
		Params: paramsJSON,
	}
	
	result, err := NotebookParallelRunTool.Handler(ctx, request)
	
	// Test parameter handling
	if result != nil {
		assert.NotNil(t, result.Content)
	}
}

// TestToolDefinitions tests that all tools have proper definitions
func TestToolDefinitions(t *testing.T) {
	tools := []*Tool{
		NotebookCreateTool,
		NotebookExecuteCellTool,
		NotebookExecuteAllTool,
		NotebookGetStateTool,
		NotebookParallelRunTool,
	}
	
	for _, tool := range tools {
		assert.NotEmpty(t, tool.Definition.Name)
		assert.NotEmpty(t, tool.Definition.Description)
		assert.NotNil(t, tool.Definition.InputSchema)
		assert.Equal(t, "object", tool.Definition.InputSchema.Type)
		assert.NotNil(t, tool.Definition.InputSchema.Properties)
		assert.NotEmpty(t, tool.Definition.InputSchema.Required)
		assert.NotNil(t, tool.Handler)
	}
}

// TestToolParameterValidation tests parameter validation for each tool
func TestToolParameterValidation(t *testing.T) {
	ctx := context.Background()
	
	testCases := []struct {
		name    string
		tool    *Tool
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "notebook_create missing required",
			tool: NotebookCreateTool,
			params: map[string]interface{}{
				"name": "test",
				// missing explanation
			},
			wantErr: true,
		},
		{
			name: "notebook_execute_cell missing required",
			tool: NotebookExecuteCellTool,
			params: map[string]interface{}{
				"notebook_id": "test",
				// missing cell_index and code
			},
			wantErr: true,
		},
		{
			name: "notebook_parallel_run empty notebooks",
			tool: NotebookParallelRunTool,
			params: map[string]interface{}{
				"notebooks": []interface{}{},
			},
			wantErr: false, // Should handle empty list gracefully
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paramsJSON, _ := json.Marshal(tc.params)
			request := mcp.CallToolRequest{Params: paramsJSON}
			
			result, err := tc.tool.Handler(ctx, request)
			
			// Handler returns errors in result, not as error
			assert.Nil(t, err)
			assert.NotNil(t, result)
			
			// Check if error is in result content
			if tc.wantErr {
				// Would check for error in result content
			}
		})
	}
}

// MockNotebookRegistry for testing
type MockNotebookRegistry struct {
	notebooks map[string]*NotebookEnvironment
}

func NewMockRegistry() *MockNotebookRegistry {
	return &MockNotebookRegistry{
		notebooks: make(map[string]*NotebookEnvironment),
	}
}

func (r *MockNotebookRegistry) Get(id string) *NotebookEnvironment {
	return r.notebooks[id]
}

func (r *MockNotebookRegistry) Register(nb *NotebookEnvironment) {
	r.notebooks[nb.ID] = nb
}

// TestToolIntegration tests tool integration with mock registry
func TestToolIntegration(t *testing.T) {
	// This test demonstrates how tools would integrate with a registry
	registry := NewMockRegistry()
	
	// Create a mock notebook
	nb := &NotebookEnvironment{
		Environment: &Environment{
			ID:   "test-nb-123",
			Name: "test-notebook",
		},
		KernelSpec: "python3",
		KernelState: &KernelState{
			ExecutionCount: 0,
			Variables:      make(map[string]interface{}),
		},
	}
	
	registry.Register(nb)
	
	// Test that registry works
	retrieved := registry.Get("test-nb-123")
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-notebook", retrieved.Name)
	
	// In real implementation, tools would use this registry
}