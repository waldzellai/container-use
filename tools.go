package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Tool struct {
	Definition mcp.Tool
	Handler    server.ToolHandlerFunc
}

var tools = []*Tool{}

func RegisterTool(tool ...*Tool) {
	tools = append(tools, tool...)
}

func init() {
	RegisterTool(
		SandboxCreateTool,
		SandboxListTool,
		RunTerminalCmdTool,
		ReadFileTool,
	)
}

var SandboxCreateTool = &Tool{
	Definition: mcp.NewTool("sandbox_create",
		mcp.WithDescription("Create a new sandbox. The sandbox only contains the base image specified, anything else required will need to be installed by hand"),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this sandbox is being created."),
			mcp.Required(),
		),
		mcp.WithString("workdir",
			mcp.Description("The local directory to be loaded in the sandbox."),
			mcp.Required(),
		),
		mcp.WithString("image",
			mcp.Description("The base image this workspace will use (e.g. alpine:latest, ubuntu:24.04, etc.)"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workdir, ok := request.GetArguments()["workdir"].(string)
		if !ok {
			return nil, errors.New("workdir must be a string")
		}
		image, ok := request.GetArguments()["image"].(string)
		if !ok {
			return nil, errors.New("image must be a string")
		}
		sandbox := CreateSandbox(image, workdir)
		return mcp.NewToolResultText(fmt.Sprintf(`{"id": %q}`, sandbox.ID)), nil
	},
}

var SandboxListTool = &Tool{
	Definition: mcp.NewTool("sandbox_list",
		mcp.WithDescription("List available sandboxes"),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this sandbox is being created."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sandboxes := ListSandboxes()
		out, err := json.Marshal(sandboxes)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(string(out)), nil
	},
}

var RunTerminalCmdTool = &Tool{
	Definition: mcp.NewTool("sandbox_run_cmd",
		mcp.WithDescription("Run a command on behalf of the user in the terminal."),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this command is being run."),
			mcp.Required(),
		),
		mcp.WithString("sandbox_id",
			mcp.Description("The identifier of the sandbox for this command. Must call `sandbox_create` first."),
			mcp.Required(),
		),
		mcp.WithString("command",
			mcp.Description("The terminal command to execute"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sandboxID, ok := request.GetArguments()["sandbox_id"].(string)
		if !ok {
			return nil, errors.New("sandbox_id must be a string")
		}
		sandbox := GetSandbox(sandboxID)
		if sandbox == nil {
			return nil, errors.New("sandbox not found")
		}
		command, ok := request.GetArguments()["command"].(string)
		if !ok {
			return nil, errors.New("command must be a string")
		}
		stdout, err := sandbox.RunCmd(ctx, command)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(stdout), nil
	},
}

var ReadFileTool = &Tool{
	Definition: mcp.NewTool("sandbox_read_file",
		mcp.WithDescription("Read the contents of a file, specifying a line range or the entire file."),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this file is being read."),
			mcp.Required(),
		),
		mcp.WithString("sandbox_id",
			mcp.Description("The identifier of the sandbox for this command. Must call `sandbox_create` first."),
			mcp.Required(),
		),
		mcp.WithString("target_file",
			mcp.Description("Path of the file to read."),
			mcp.Required(),
		),
		mcp.WithBoolean("should_read_entire_file",
			mcp.Description("Whether to read the entire file. Defaults to false."),
		),
		mcp.WithNumber("start_line_one_indexed",
			mcp.Description("The one-indexed line number to start reading from (inclusive)."),
		),
		mcp.WithNumber("end_line_one_indexed_inclusive",
			mcp.Description("The one-indexed line number to end reading at (inclusive)."),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sandboxID, ok := request.GetArguments()["sandbox_id"].(string)
		if !ok {
			return nil, errors.New("sandbox_id must be a string")
		}
		sandbox := GetSandbox(sandboxID)
		if sandbox == nil {
			return nil, errors.New("sandbox not found")
		}

		targetFile, ok := request.GetArguments()["target_file"].(string)
		if !ok {
			return nil, errors.New("target_file must be a string")
		}
		shouldReadEntireFile, _ := request.GetArguments()["should_read_entire_file"].(bool)
		startLineOneIndexed, _ := request.GetArguments()["start_line_one_indexed"].(int)
		endLineOneIndexedInclusive, _ := request.GetArguments()["end_line_one_indexed_inclusive"].(int)

		fileContents, err := sandbox.ReadFile(ctx, targetFile, shouldReadEntireFile, startLineOneIndexed, endLineOneIndexedInclusive)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fileContents), nil
	},
}
