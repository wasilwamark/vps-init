package cli

import (
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

func InitPluginSystem() error {
	registry := plugin.GetBuiltinRegistry()
	_ = registry // Use registry to avoid unused import error
	return nil
}

func Execute() error {
	return rootCmd.Execute()
}