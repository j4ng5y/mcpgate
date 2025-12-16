package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/j4ng5y/mcpgate/config"
	"github.com/j4ng5y/mcpgate/mcp"
	"github.com/j4ng5y/mcpgate/server"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run mcpgate as an MCP server",
	Long: `Start mcpgate as a Model Context Protocol server.

The server reads JSON-RPC 2.0 requests from stdin and writes responses to stdout.
It routes requests to configured upstream MCP servers.`,
	Run: runServer,
}

func init() {
	serverCmd.Flags().StringVarP(&configPath, "config", "c", "config.toml", "Path to configuration file")
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize server manager
	mgr := server.NewManager(cfg)
	if err := mgr.Start(); err != nil {
		log.Fatalf("Failed to start server manager: %v", err)
	}

	// Create MCP router
	router := mcp.NewRouter(mgr)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		mgr.Stop()
		cancel()
		os.Exit(0)
	}()

	// Start stdio server
	reader := bufio.NewReader(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			break
		}

		var request mcp.Request
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			// Send error response
			errResp := mcp.Response{
				JSONRPC: "2.0",
				Error: &mcp.JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			if err := encoder.Encode(errResp); err != nil {
				log.Printf("Error encoding error response: %v", err)
			}
			continue
		}

		// Route request
		response := router.Route(ctx, &request)
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}
