package cli

import (
	"os"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

func InitPluginSystem() error {
	registry := plugin.GetBuiltinRegistry()
	_ = registry // Use registry to avoid unused import error
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
		// Check aliases and help
		if cmdName == "help" || cmdName == "--help" || cmdName == "-h" {
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
