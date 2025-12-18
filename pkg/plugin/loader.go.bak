package plugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config represents plugin configuration
type Config struct {
	Plugins map[string]PluginConfig `yaml:"plugins" json:"plugins"`
	Paths   []string                `yaml:"paths" json:"paths"`
}

// PluginConfig represents individual plugin configuration
type PluginConfig struct {
	Enabled bool                   `yaml:"enabled" json:"enabled"`
	Path    string                 `yaml:"path" json:"path"`
	Config  map[string]interface{} `yaml:"config" json:"config"`
	Import  string                 `yaml:"import" json:"import"`
	Remote  *RemotePlugin          `yaml:"remote" json:"remote"`
}

// RemotePlugin represents a remote plugin configuration
type RemotePlugin struct {
	URL    string `yaml:"url" json:"url"`
	SHA256 string `yaml:"sha256" json:"sha256"`
}

// FSLoader loads plugins from the filesystem
type FSLoader struct {
	configPath string
	paths      []string
	config     Config
}

// NewFSLoader creates a new filesystem plugin loader
func NewFSLoader(configPath string) (*FSLoader, error) {
	loader := &FSLoader{
		configPath: configPath,
		paths: []string{
			"./plugins",
			"/usr/local/lib/vps-init/plugins",
			"/opt/vps-init/plugins",
		},
	}

	// Load configuration
	if err := loader.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load plugin config: %w", err)
	}

	// Add custom paths from config
	loader.paths = append(loader.paths, loader.config.Paths...)

	return loader, nil
}

// loadConfig loads the plugin configuration
func (l *FSLoader) loadConfig() error {
	// Default config
	l.config = Config{
		Plugins: make(map[string]PluginConfig),
		Paths:   []string{},
	}

	// Try to load from file
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		// Create default config file
		defaultConfig := `# VPS-Init Plugin Configuration
plugins:
  nginx:
    enabled: true
    import: "github.com/wasilwamark/vps-init-plugins/nginx"
  docker:
    enabled: true
    import: "github.com/wasilwamark/vps-init-plugins/docker"
  monitoring:
    enabled: true
    import: "github.com/wasilwamark/vps-init-plugins/monitoring"

# Additional plugin search paths
paths:
  - "./plugins"
  - "~/.vps-init/plugins"
  - "/usr/local/lib/vps-init/plugins"
`
		os.WriteFile(l.configPath, []byte(defaultConfig), 0644)
		return nil
	}

	data, err := ioutil.ReadFile(l.configPath)
	if err != nil {
		return err
	}

	// Try YAML first
	if err := yaml.Unmarshal(data, &l.config); err != nil {
		// Fall back to JSON
		return json.Unmarshal(data, &l.config)
	}

	return nil
}

// LoadPlugins loads all enabled plugins
func (l *FSLoader) LoadPlugins() ([]Plugin, error) {
	var plugins []Plugin

	for name, config := range l.config.Plugins {
		if !config.Enabled {
			continue
		}

		plugin, err := l.loadPluginByName(name, config)
		if err != nil {
			fmt.Printf("Warning: failed to load plugin %s: %v\n", name, err)
			continue
		}

		plugins = append(plugins, plugin)
	}

	// Also discover plugins in paths
	discovered, err := l.discoverPlugins()
	if err != nil {
		return nil, err
	}

	plugins = append(plugins, discovered...)

	return plugins, nil
}

// LoadPlugin loads a specific plugin by name
func (l *FSLoader) LoadPlugin(name string) (Plugin, error) {
	config, exists := l.config.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found in configuration", name)
	}

	return l.loadPluginByName(name, config)
}

// loadPluginByName loads a plugin by name and config
func (l *FSLoader) loadPluginByName(name string, config PluginConfig) (Plugin, error) {
	// Method 1: Direct path
	if config.Path != "" {
		return l.loadFromPath(config.Path, config.Config)
	}

	// Method 2: Import (Go package)
	if config.Import != "" {
		return l.loadFromImport(config.Import, config.Config)
	}

	// Method 3: Remote plugin
	if config.Remote != nil {
		return l.loadFromRemote(config.Remote, config.Config)
	}

	// Method 4: Search in paths
	return l.discoverPlugin(name)
}

// loadFromPath loads a plugin from a file path
func (l *FSLoader) loadFromPath(path string, config map[string]interface{}) (Plugin, error) {
	// Support both .so files and Go packages
	if strings.HasSuffix(path, ".so") {
		return l.loadFromSharedObject(path, config)
	}

	// Assume it's a Go package path
	return l.loadFromImport(path, config)
}

// loadFromImport loads a plugin from a Go import path
func (l *FSLoader) loadFromImport(importPath string, config map[string]interface{}) (Plugin, error) {
	// This would be used in a compiled-in plugin scenario
	// For now, we'll implement a simple registry approach
	registry := GetBuiltinRegistry()
	if plugin, exists := registry.Get(importPath); exists {
		if err := plugin.Initialize(config); err != nil {
			return nil, err
		}
		return plugin, nil
	}

	return nil, fmt.Errorf("builtin plugin not found: %s", importPath)
}

// loadFromSharedObject loads a plugin from a .so file
func (l *FSLoader) loadFromSharedObject(path string, config map[string]interface{}) (Plugin, error) {
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	// Look for NewPlugin symbol
	newPluginSymbol, err := plug.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin %s does not export NewPlugin: %w", path, err)
	}

	// Type assert to function
	newPlugin, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin %s NewPlugin has wrong signature", path)
	}

	pluginInstance := newPlugin()
	if err := pluginInstance.Initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin %s: %w", path, err)
	}

	return pluginInstance, nil
}

// loadFromRemote loads a plugin from a remote URL
func (l *FSLoader) loadFromRemote(remote *RemotePlugin, config map[string]interface{}) (Plugin, error) {
	// TODO: Implement remote plugin downloading
	return nil, fmt.Errorf("remote plugins not yet implemented")
}

// discoverPlugins discovers plugins in configured paths
func (l *FSLoader) discoverPlugins() ([]Plugin, error) {
	var plugins []Plugin

	for _, path := range l.paths {
		// Expand ~
		expandedPath := os.ExpandEnv(path)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		// Check if path exists before walking
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			continue
		}

		// Walk directory
		err := filepath.Walk(expandedPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Look for .so files
			if strings.HasSuffix(path, ".so") && !info.IsDir() {
				plugin, err := l.loadFromSharedObject(path, nil)
				if err != nil {
					fmt.Printf("Warning: failed to load plugin %s: %v\n", path, err)
					return nil
				}
				plugins = append(plugins, plugin)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Warning: failed to walk plugin path %s: %v\n", path, err)
		}
	}

	return plugins, nil
}

// discoverPlugin discovers a specific plugin by name
func (l *FSLoader) discoverPlugin(name string) (Plugin, error) {
	for _, path := range l.paths {
		expandedPath := os.ExpandEnv(path)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		// Look for .so files
		soPath := filepath.Join(expandedPath, name+".so")
		if _, err := os.Stat(soPath); err == nil {
			return l.loadFromSharedObject(soPath, nil)
		}

		// Look for directory with plugin.yaml
		dirPath := filepath.Join(expandedPath, name)
		if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
			// Load plugin.yaml
			configPath := filepath.Join(dirPath, "plugin.yaml")
			data, err := ioutil.ReadFile(configPath)
			if err != nil {
				continue
			}

			var pluginMeta PluginMetadata
			if err := yaml.Unmarshal(data, &pluginMeta); err != nil {
				continue
			}

			// Look for .so file in directory
			soPath = filepath.Join(dirPath, name+".so")
			if _, err := os.Stat(soPath); err == nil {
				return l.loadFromSharedObject(soPath, pluginMeta.Config)
			}
		}
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

// ListAvailablePlugins lists all available plugins
func (l *FSLoader) ListAvailablePlugins() ([]PluginMetadata, error) {
	var plugins []PluginMetadata

	// Add configured plugins
	for name, config := range l.config.Plugins {
		meta := PluginMetadata{
			Name: name,
		}
		if config.Config != nil {
			meta.Config = config.Config
		}
		plugins = append(plugins, meta)
	}

	// Discover plugins in paths
	for _, path := range l.paths {
		expandedPath := os.ExpandEnv(path)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		filepath.Walk(expandedPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			// Look for plugin.yaml files
			if strings.HasSuffix(path, "plugin.yaml") {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return nil
				}

				var meta PluginMetadata
				if err := yaml.Unmarshal(data, &meta); err != nil {
					return nil
				}

				plugins = append(plugins, meta)
			}

			return nil
		})
	}

	return plugins, nil
}
