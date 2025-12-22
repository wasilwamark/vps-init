package cli

import (
	"os"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

func InitPluginSystem() error {
	// Initialize built-in plugin loader
	loader := plugin.NewLoader()

	// Get registry and set loader
	registry := plugin.GetBuiltinRegistry()
	registry.SetLoader(loader)

	// Load all plugins (this will just register built-in plugins)
	if err := registry.LoadAll(); err != nil {
		// Log warning but don't fail
		// fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	return nil
}

func Execute() error {
	// Load commands from plugins
	registry := plugin.GetBuiltinRegistry()
	for _, cmd := range registry.GetRootCommands() {
		rootCmd.AddCommand(cmd)
	}

	// Check if the first argument is a known command
	if len(os.Args) > 1 {
		cmdName := os.Args[1]
		// Check aliases, help, and version
		if cmdName == "help" || cmdName == "--help" || cmdName == "-h" || cmdName == "--version" || cmdName == "-v" {
			return rootCmd.Execute()
		}

		// Check registered commands
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == cmdName || cmd.HasAlias(cmdName) {
				found = true
				break
			}
		}

		// If not a known command, assume direct execution mode
		if !found {
			executeDirectCommand()
			return nil
		}
	}

	return rootCmd.Execute()
}
