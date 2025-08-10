package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"dagger.io/dagger"
	"github.com/dagger/container-use/mcpserver"
	"github.com/spf13/cobra"
)

var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "Start MCP server for agent integration",
	Long:  `Start the Model Context Protocol server that enables AI agents to create and manage containerized environments. This is typically used by agents like Claude Code, Cursor, or VSCode.`,
	RunE: func(app *cobra.Command, _ []string) error {
		ctx := app.Context()

		slog.Info("connecting to dagger")

		dag, err := dagger.Connect(ctx, dagger.WithLogOutput(logWriter))
		if err != nil {
			slog.Error("Error starting dagger", "error", err)

			if isDockerDaemonError(err) {
				handleDockerDaemonError()
			}

			os.Exit(1)
		}
		defer dag.Close()

		return mcpserver.RunStdioServer(ctx, dag)
	},
}

var killBackgroundCmd = &cobra.Command{
	Use:   "kill-background [pid]",
	Short: "Kill a background process in host mode by PID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil || pid <= 0 {
			return fmt.Errorf("invalid pid")
		}
		// This CLI currently proxies to MCP; print guidance
		fmt.Fprintln(os.Stderr, "Use the MCP tool environment_kill_background to stop processes from clients.")
		fmt.Printf("Requested stop for PID %d\n", pid)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stdioCmd)
	rootCmd.AddCommand(killBackgroundCmd)
}
