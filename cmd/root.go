package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcpgate",
	Short: "MCP Gateway - Route requests between MCP clients and servers",
	Long: `mcpgate is a Model Context Protocol (MCP) Gateway that routes JSON-RPC requests
between MCP clients and upstream MCP servers via various transport methods.

It acts as a local MCP server on stdout and supports configuration of multiple
upstream servers via different transports (stdio, HTTP, WebSocket, Unix sockets).`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(injectCmd)
}
