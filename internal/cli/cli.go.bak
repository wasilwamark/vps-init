package cli

import (
	"os"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

func InitPluginSystem() error {
	// Initialize FS Loader
	loader, err := plugin.NewFSLoader(os.ExpandEnv("$HOME/.vps-init/plugins/config.yaml"))
	if err != nil {
		return err
	}

	// Get registry (builtin + fs)
	registry := plugin.GetBuiltinRegistry()
	registry.SetLoader(loader)

	// Load all plugins
	if err := registry.LoadAll(); err != nil {
		// Log warning but don't fail, maybe just no plugins found
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
