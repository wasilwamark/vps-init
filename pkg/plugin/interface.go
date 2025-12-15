package plugin

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Metadata
	Name() string
	Description() string
	Version() string
	Author() string

	// Initialization
	Initialize(config map[string]interface{}) error
	Validate() error

	// Commands
	GetCommands() []Command
	GetRootCommand() *cobra.Command

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Dependencies
	Dependencies() []Dependency
	Compatibility() Compatibility
	GetMetadata() PluginMetadata
}

// Dependency represents a plugin dependency
type Dependency struct {
	Name     string   `json:"name"`
	Version  string   `json:"version,omitempty"`  // Semantic version constraint
	Optional bool     `json:"optional"`
	Tags     []string `json:"tags,omitempty"`
}

// Compatibility defines plugin compatibility requirements
type Compatibility struct {
	MinVPSInitVersion string `json:"min_vps_init_version"`
	MaxVPSInitVersion string `json:"max_vps_init_version,omitempty"`
	GoVersion         string `json:"go_version,omitempty"`
	Platforms         []string `json:"platforms,omitempty"` // e.g., ["linux/amd64", "linux/arm64"]
	Tags              []string `json:"tags,omitempty"`
}

// Command represents a plugin command
type Command struct {
	Name        string
	Description string
	Aliases     []string
	Args        []Argument
	Flags       []Flag
	Handler     CommandHandler
}

// Argument defines a command argument
type Argument struct {
	Name        string
	Description string
	Required    bool
	Type        ArgumentType
}

// ArgumentType represents the type of argument
type ArgumentType int

const (
	ArgumentTypeString ArgumentType = iota
	ArgumentTypeInt
	ArgumentTypeBool
	ArgumentTypeSlice
)

// Flag defines a command flag
type Flag struct {
	Name        string
	Shorthand   string
	Description string
	Default     interface{}
	Required    bool
	Type        ArgumentType
}

// CommandHandler handles command execution
type CommandHandler func(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Version           string                 `json:"version"`
	Author            string                 `json:"author"`
	License           string                 `json:"license"`
	Homepage          string                 `json:"homepage"`
	Repository        string                 `json:"repository"`
	Tags              []string               `json:"tags"`
	Config            map[string]interface{} `json:"config,omitempty"`

	// Installation information
	InstallPath       string                 `json:"install_path,omitempty"`
	InstalledAt       string                 `json:"installed_at,omitempty"`
	LastUpdated       string                 `json:"last_updated,omitempty"`
	Checksum          string                 `json:"checksum,omitempty"`
	Source            string                 `json:"source,omitempty"`        // git URL, package path
	BuildInfo         BuildInfo              `json:"build_info,omitempty"`

	// Validation information
	Validated         bool                   `json:"validated"`
	ValidationErrors  []string               `json:"validation_errors,omitempty"`
	Signature         string                 `json:"signature,omitempty"`     // GPG signature
	TrustLevel        string                 `json:"trust_level,omitempty"`  // official, community, untrusted
}

// BuildInfo contains build-related plugin information
type BuildInfo struct {
	GoVersion    string   `json:"go_version,omitempty"`
	BuildTime    string   `json:"build_time,omitempty"`
	GitCommit    string   `json:"git_commit,omitempty"`
	GitTag       string   `json:"git_tag,omitempty"`
	BuildFlags   []string `json:"build_flags,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// PluginLoader defines how plugins are loaded
type PluginLoader interface {
	LoadPlugins() ([]Plugin, error)
	LoadPlugin(name string) (Plugin, error)
	ListAvailablePlugins() ([]PluginMetadata, error)
}

// Registry manages loaded plugins
type Registry struct {
	plugins map[string]Plugin
	loader  PluginLoader
}

// NewRegistry creates a new plugin registry
func NewRegistry(loader PluginLoader) *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		loader:  loader,
	}
}

// SetLoader sets the plugin loader
func (r *Registry) SetLoader(loader PluginLoader) {
	r.loader = loader
}

// Register registers a plugin
func (r *Registry) Register(plugin Plugin) {
	r.plugins[plugin.Name()] = plugin
}

// Get gets a plugin by name
func (r *Registry) Get(name string) (Plugin, bool) {
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// GetAll returns all registered plugins
func (r *Registry) GetAll() []Plugin {
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// LoadAll loads all available plugins
func (r *Registry) LoadAll() error {
	plugins, err := r.loader.LoadPlugins()
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		r.Register(plugin)
	}

	return nil
}

// LoadPlugin loads a specific plugin
func (r *Registry) LoadPlugin(name string) error {
	plugin, err := r.loader.LoadPlugin(name)
	if err != nil {
		return err
	}

	r.Register(plugin)
	return nil
}

// GetCommands returns all commands from all plugins
func (r *Registry) GetCommands() []Command {
	var commands []Command
	for _, plugin := range r.plugins {
		commands = append(commands, plugin.GetCommands()...)
	}
	return commands
}

// GetRootCommands returns root commands from all plugins
func (r *Registry) GetRootCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, plugin := range r.plugins {
		if cmd := plugin.GetRootCommand(); cmd != nil {
			commands = append(commands, cmd)
		}
	}
	return commands
}
