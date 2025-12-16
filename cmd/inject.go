package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/j4ng5y/mcpgate/inject"
	"github.com/spf13/cobra"
)

var (
	injectURL      string
	injectName     string
	injectAgents   string
	injectMode     string
	injectConfig   string
	doEject        bool
)

// injectCmd represents the inject command
var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject mcpgate into AI agent configurations",
	Long: `Inject or remove mcpgate from various AI agent configurations.

This command automatically finds installed AI agents and adds mcpgate as an MCP server.
It creates backups of agent configs before modification for safe recovery.

Supported agents:
  - Claude Desktop (local configuration)
  - Cursor (local configuration)
  - Zed (local configuration)
  - Gemini CLI (local configuration)
  - Codex CLI (local configuration)
  - OpenCode (local configuration)
  - Windsurf (local configuration)
  - Kiro (local configuration)`,
	Run: runInject,
}

func init() {
	injectCmd.Flags().StringVar(&injectMode, "mode", "stdio", "Connection mode: stdio (subprocess) or http (HTTP server)")
	injectCmd.Flags().StringVar(&injectURL, "url", "", "URL to the mcpgate server (HTTP mode only)")
	injectCmd.Flags().StringVar(&injectName, "name", "mcpgate", "Name for the mcpgate server entry")
	injectCmd.Flags().StringVar(&injectAgents, "agents", "all", "Comma-separated list of agents to inject into (all, claude, cursor, zed, codex-cli, gemini-cli, opencode, windsurf, kiro)")
	injectCmd.Flags().StringVar(&injectConfig, "config", "", "Path to mcpgate config file (stdio mode only)")
	injectCmd.Flags().BoolVar(&doEject, "eject", false, "Remove mcpgate from agent configs instead of injecting")
}

func runInject(cmd *cobra.Command, args []string) {
	// Validate mode
	if injectMode != "stdio" && injectMode != "http" {
		fmt.Printf("Error: invalid mode '%s'. Must be 'stdio' or 'http'\n", injectMode)
		return
	}

	// Validate mode-specific parameters
	if injectMode == "stdio" {
		// For stdio mode, find mcpgate binary
		exe, err := os.Executable()
		if err != nil {
			fmt.Printf("Error: failed to find mcpgate binary: %v\n", err)
			return
		}

		// Build args for mcpgate subprocess
		var args []string
		if injectConfig != "" {
			args = []string{"server", "-c", injectConfig}
		} else {
			args = []string{"server"}
		}

		// Create manager and register agents
		manager := inject.NewManager()
		manager.RegisterAgent(inject.NewClaude())
		manager.RegisterAgent(inject.NewCursor())
		manager.RegisterAgent(inject.NewZed())
		manager.RegisterAgent(inject.NewCodexCLI())
		manager.RegisterAgent(inject.NewGeminiCLI())
		manager.RegisterAgent(inject.NewOpenCode())
		manager.RegisterAgent(inject.NewWindsurf())
		manager.RegisterAgent(inject.NewKiro())

		if doEject {
			handleEject(manager)
		} else {
			handleInjectStdio(manager, exe, args)
		}
	} else {
		// HTTP mode
		if injectURL == "" {
			fmt.Println("Error: --url is required for HTTP mode")
			return
		}

		// Create manager and register agents
		manager := inject.NewManager()
		manager.RegisterAgent(inject.NewClaude())
		manager.RegisterAgent(inject.NewCursor())
		manager.RegisterAgent(inject.NewZed())
		manager.RegisterAgent(inject.NewCodexCLI())
		manager.RegisterAgent(inject.NewGeminiCLI())
		manager.RegisterAgent(inject.NewOpenCode())
		manager.RegisterAgent(inject.NewWindsurf())
		manager.RegisterAgent(inject.NewKiro())

		if doEject {
			handleEject(manager)
		} else {
			handleInjectHTTP(manager)
		}
	}
}

// handleInjectStdio injects mcpgate (stdio mode) into agent configs
func handleInjectStdio(manager *inject.Manager, command string, args []string) {
	installed := manager.ListInstalledAgents()

	if len(installed) == 0 {
		fmt.Println("No supported agents found installed on this system.")
		fmt.Println("\nSupported agents:")
		fmt.Println("  - Claude Desktop")
		fmt.Println("  - Cursor")
		fmt.Println("  - Zed")
		fmt.Println("  - Gemini CLI")
		fmt.Println("  - Codex CLI")
		fmt.Println("  - OpenCode")
		return
	}

	fmt.Printf("Found %d installed agent(s).\n\n", len(installed))

	var agentsToInject []inject.Agent

	if injectAgents == "all" {
		agentsToInject = installed
	} else {
		agentNames := parseAgentList(injectAgents)
		for _, agent := range installed {
			for _, name := range agentNames {
				if isAgentMatch(agent.Name(), name) {
					agentsToInject = append(agentsToInject, agent)
					break
				}
			}
		}
	}

	if len(agentsToInject) == 0 {
		fmt.Println("No matching agents found.")
		return
	}

	fmt.Printf("Injecting mcpgate (stdio mode) into %d agent(s)...\n", len(agentsToInject))
	fmt.Printf("Command: %s %v\n\n", command, args)

	options := map[string]interface{}{}

	for _, agent := range agentsToInject {
		fmt.Printf("  Injecting into %s... ", agent.Name())

		if err := agent.CreateBackup(); err != nil {
			fmt.Printf("FAILED (backup error: %v)\n", err)
			log.Printf("Failed to backup %s: %v", agent.Name(), err)
			continue
		}

		if err := agent.InjectStdio(command, args, injectName, options); err != nil {
			fmt.Printf("FAILED (%v)\n", err)
			log.Printf("Failed to inject into %s: %v", agent.Name(), err)
			if restoreErr := agent.RestoreBackup(); restoreErr != nil {
				fmt.Printf("    WARNING: Failed to restore backup: %v\n", restoreErr)
			}
			continue
		}

		fmt.Println("OK")
	}

	fmt.Printf("\nSuccessfully injected mcpgate (Name: %s)\n", injectName)
}

// handleInjectHTTP injects mcpgate (HTTP mode) into agent configs
func handleInjectHTTP(manager *inject.Manager) {
	installed := manager.ListInstalledAgents()

	if len(installed) == 0 {
		fmt.Println("No supported agents found installed on this system.")
		fmt.Println("\nSupported agents:")
		fmt.Println("  - Claude Desktop")
		fmt.Println("  - Cursor")
		fmt.Println("  - Zed")
		fmt.Println("  - Gemini CLI")
		fmt.Println("  - Codex CLI")
		fmt.Println("  - OpenCode")
		return
	}

	fmt.Printf("Found %d installed agent(s).\n\n", len(installed))

	var agentsToInject []inject.Agent

	if injectAgents == "all" {
		agentsToInject = installed
	} else {
		agentNames := parseAgentList(injectAgents)
		for _, agent := range installed {
			for _, name := range agentNames {
				if isAgentMatch(agent.Name(), name) {
					agentsToInject = append(agentsToInject, agent)
					break
				}
			}
		}
	}

	if len(agentsToInject) == 0 {
		fmt.Println("No matching agents found.")
		return
	}

	fmt.Printf("Injecting mcpgate (HTTP mode) into %d agent(s)...\n", len(agentsToInject))
	fmt.Printf("URL: %s\n\n", injectURL)

	options := map[string]interface{}{}

	for _, agent := range agentsToInject {
		fmt.Printf("  Injecting into %s... ", agent.Name())

		if err := agent.CreateBackup(); err != nil {
			fmt.Printf("FAILED (backup error: %v)\n", err)
			log.Printf("Failed to backup %s: %v", agent.Name(), err)
			continue
		}

		if err := agent.InjectHTTP(injectURL, injectName, options); err != nil {
			fmt.Printf("FAILED (%v)\n", err)
			log.Printf("Failed to inject into %s: %v", agent.Name(), err)
			if restoreErr := agent.RestoreBackup(); restoreErr != nil {
				fmt.Printf("    WARNING: Failed to restore backup: %v\n", restoreErr)
			}
			continue
		}

		fmt.Println("OK")
	}

	fmt.Printf("\nSuccessfully injected mcpgate (URL: %s, Name: %s)\n", injectURL, injectName)
}

// handleEject removes mcpgate from agent configs
func handleEject(manager *inject.Manager) {
	injected := manager.ListInjectedAgents(injectName)

	if len(injected) == 0 {
		fmt.Printf("mcpgate '%s' is not injected into any installed agents.\n", injectName)
		return
	}

	fmt.Printf("Found %d agent(s) with mcpgate '%s' injected.\n\n", len(injected), injectName)
	fmt.Printf("Removing mcpgate from %d agent(s)...\n\n", len(injected))

	for _, agent := range injected {
		fmt.Printf("  Removing from %s... ", agent.Name())

		if err := agent.Eject(injectName); err != nil {
			fmt.Printf("FAILED (%v)\n", err)
			log.Printf("Failed to eject from %s: %v", agent.Name(), err)
			continue
		}

		fmt.Println("OK")
	}

	fmt.Printf("\nSuccessfully removed mcpgate '%s' from all agents\n", injectName)
}

// parseAgentList parses a comma-separated list of agent names
func parseAgentList(agents string) []string {
	var result []string
	// Simple split by comma
	var current string
	for _, c := range agents {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// isAgentMatch checks if an agent name matches a given identifier
func isAgentMatch(agentName, identifier string) bool {
	matches := map[string][]string{
		"claude":     {"Claude Desktop", "claude"},
		"cursor":     {"Cursor", "cursor"},
		"zed":        {"Zed", "zed"},
		"codex-cli":  {"Codex CLI", "codex-cli", "codex"},
		"gemini-cli": {"Gemini CLI", "gemini-cli", "gemini"},
		"opencode":   {"OpenCode", "opencode"},
		"windsurf":   {"Windsurf", "windsurf"},
		"kiro":       {"Kiro", "kiro"},
	}

	if names, ok := matches[identifier]; ok {
		for _, name := range names {
			if agentName == name {
				return true
			}
		}
	}

	return agentName == identifier
}
