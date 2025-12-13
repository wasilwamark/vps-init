package plugin

import (
	"sync"
)

// builtinRegistry holds built-in plugins
var builtinRegistry = &struct {
	sync.RWMutex
	plugins map[string]Plugin
}{
	plugins: make(map[string]Plugin),
}

// RegisterBuiltin registers a built-in plugin
func RegisterBuiltin(importPath string, plugin Plugin) {
	builtinRegistry.Lock()
	defer builtinRegistry.Unlock()
	builtinRegistry.plugins[importPath] = plugin
}

// GetBuiltinRegistry returns the built-in plugin registry
func GetBuiltinRegistry() *Registry {
	builtinRegistry.RLock()
	defer builtinRegistry.RUnlock()

	registry := NewRegistry(&dummyLoader{})

	// Register all built-in plugins
	for _, plugin := range builtinRegistry.plugins {
		registry.Register(plugin)
	}

	return registry
}

// dummyLoader is used for built-in registry
type dummyLoader struct{}

func (d *dummyLoader) LoadPlugins() ([]Plugin, error) {
	return nil, nil
}

func (d *dummyLoader) LoadPlugin(name string) (Plugin, error) {
	return nil, nil
}

func (d *dummyLoader) ListAvailablePlugins() ([]PluginMetadata, error) {
	return nil, nil
}