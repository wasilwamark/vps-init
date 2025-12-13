package main

import (
	"fmt"
	"os"

	"github.com/wasilwamark/vps-init/internal/cli"
)

func main() {
	// Initialize built-in plugins first
	initializeBuiltinPlugins()

	// Initialize plugin system first
	if err := cli.InitPluginSystem(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize plugins: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}