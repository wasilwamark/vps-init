package plugin

import (
	"fmt"
)

// Loader manages loading of built-in plugins
type Loader struct {
	registry *Registry
}

// NewLoader creates a new plugin loader for built-in plugins
func NewLoader() *Loader {
	return &Loader{
		registry: GetBuiltinRegistry(),
	}
}

// LoadPlugins loads all built-in plugins
func (l *Loader) LoadPlugins() ([]Plugin, error) {
	return l.registry.GetAll(), nil
}

// LoadPlugin loads a specific built-in plugin by name
func (l *Loader) LoadPlugin(name string) (Plugin, error) {
	plugin, exists := l.registry.Get(name)
	if !exists {
		return nil, fmt.Errorf("builtin plugin '%s' not found", name)
	}
	return plugin, nil
}

// ListAvailablePlugins lists all available built-in plugins
func (l *Loader) ListAvailablePlugins() ([]PluginMetadata, error) {
	plugins := l.registry.GetAll()
	var metadata []PluginMetadata

	for _, p := range plugins {
		metadata = append(metadata, p.GetMetadata())
	}

	return metadata, nil
}
