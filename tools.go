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
		ContainerCreateTool,
		ContainerListTool,
		ContainerRunCmdTool,
		ContainerReadFileTool,
	)
}

var ContainerCreateTool = &Tool{
	Definition: mcp.NewTool("container_create",
		mcp.WithDescription("Create a new container. The sandbox only contains the base image specified, anything else required will need to be installed by hand"),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this sandbox is being created."),
			mcp.Required(),
		),
		mcp.WithString("local_workdir",
			mcp.Description("The local directory to be loaded in the sandbox."),
			mcp.Required(),
		),
		mcp.WithString("image",
			mcp.Description("The base image this workspace will use (e.g. alpine:latest, ubuntu:24.04, etc.)"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workdir, ok := request.GetArguments()["local_workdir"].(string)
		if !ok {
			return nil, errors.New("workdir must be a string")
		}
		image, ok := request.GetArguments()["image"].(string)
		if !ok {
			return nil, errors.New("image must be a string")
		}
		sandbox := CreateContainer(image, workdir)
		return mcp.NewToolResultText(fmt.Sprintf(`{"id": %q}`, sandbox.ID)), nil
	},
}

var ContainerListTool = &Tool{
	Definition: mcp.NewTool("container_list",
		mcp.WithDescription("List available containers"),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this container is being listed."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		containers := ListContainers()
		out, err := json.Marshal(containers)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(string(out)), nil
	},
}

var ContainerRunCmdTool = &Tool{
	Definition: mcp.NewTool("container_run_cmd",
		mcp.WithDescription("Run a command on behalf of the user in the terminal."),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this command is being run."),
			mcp.Required(),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container for this command. Must call `container_create` first."),
			mcp.Required(),
		),
		mcp.WithString("command",
			mcp.Description("The terminal command to execute"),
			mcp.Required(),
		),
		mcp.WithString("shell",
			mcp.Description("The shell that will be interpreting this command (default: sh)"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		containerID, ok := request.GetArguments()["container_id"].(string)
		if !ok {
			return nil, errors.New("container_id must be a string")
		}
		container := GetContainer(containerID)
		if container == nil {
			return nil, errors.New("container not found")
		}
		command, ok := request.GetArguments()["command"].(string)
		if !ok {
			return nil, errors.New("command must be a string")
		}
		shell, ok := request.GetArguments()["shell"].(string)
		if !ok {
			shell = "bash"
		}
		stdout, err := container.RunCmd(ctx, command, shell)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(stdout), nil
	},
}

var ContainerReadFileTool = &Tool{
	Definition: mcp.NewTool("container_read_file",
		mcp.WithDescription("Read the contents of a file, specifying a line range or the entire file."),
		mcp.WithString("explanation",
			mcp.Description("One sentence explanation for why this file is being read."),
			mcp.Required(),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container for this command. Must call `container_create` first."),
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
		containerID, ok := request.GetArguments()["container_id"].(string)
		if !ok {
			return nil, errors.New("container_id must be a string")
		}
		container := GetContainer(containerID)
		if container == nil {
			return nil, errors.New("container not found")
		}

		targetFile, ok := request.GetArguments()["target_file"].(string)
		if !ok {
			return nil, errors.New("target_file must be a string")
		}
		shouldReadEntireFile, _ := request.GetArguments()["should_read_entire_file"].(bool)
		startLineOneIndexed, _ := request.GetArguments()["start_line_one_indexed"].(int)
		endLineOneIndexedInclusive, _ := request.GetArguments()["end_line_one_indexed_inclusive"].(int)

		fileContents, err := container.ReadFile(ctx, targetFile, shouldReadEntireFile, startLineOneIndexed, endLineOneIndexedInclusive)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fileContents), nil
	},
}
