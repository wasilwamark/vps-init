package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Metadata
	Name() string
	Description() string
	Version() string
	Author() string

	// Initialization and Validation
	Initialize(config map[string]interface{}) error
	Validate() error

	// Commands
	GetCommands() []Command
	GetRootCommand() *cobra.Command

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Dependencies and Compatibility
	Dependencies() []Dependency
	Compatibility() Compatibility
	GetMetadata() PluginMetadata
}

// Dependency represents a plugin dependency
type Dependency struct {
	Name     string   `json:"name"`
	Version  string   `json:"version,omitempty"` // Semantic version constraint
	Optional bool     `json:"optional"`
	Tags     []string `json:"tags,omitempty"`
}

// Compatibility defines plugin compatibility requirements
type Compatibility struct {
	MinVPSInitVersion string   `json:"min_vps_init_version"`
	MaxVPSInitVersion string   `json:"max_vps_init_version,omitempty"`
	GoVersion         string   `json:"go_version,omitempty"`
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
type CommandHandler func(ctx context.Context, conn Connection, args []string, flags map[string]interface{}) error

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Homepage    string                 `json:"homepage"`
	Repository  string                 `json:"repository"`
	Tags        []string               `json:"tags"`
	Config      map[string]interface{} `json:"config,omitempty"`

	// Installation information
	InstallPath string    `json:"install_path,omitempty"`
	InstalledAt string    `json:"installed_at,omitempty"`
	LastUpdated string    `json:"last_updated,omitempty"`
	Checksum    string    `json:"checksum,omitempty"`
	Source      string    `json:"source,omitempty"` // git URL, package path
	BuildInfo   BuildInfo `json:"build_info,omitempty"`

	// Validation information
	Validated        bool     `json:"validated"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	Signature        string   `json:"signature,omitempty"`   // GPG signature
	TrustLevel       string   `json:"trust_level,omitempty"` // official, community, untrusted
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

// Connection interface for SSH connections (enhanced with all methods from vps-init-ssh)
type Connection interface {
	// Basic operations
	RunCommand(cmd string, sudo bool) Result
	RunCommandWithOutput(cmd string, sudo bool) (string, error)
	UploadFile(localPath, remotePath string) error
	DownloadFile(remotePath, localPath string) error
	Close() error

	// Enhanced SSH operations (from vps-init-ssh)
	Connect() bool
	Disconnect()
	Reconnect() error
	IsHealthy() bool
	GetConnectionStats() *ConnectionStats

	// Command execution
	RunSudo(cmd, password string) Result
	RunInteractive(cmd string) error
	Shell() error

	// File operations
	WriteFile(content, path string) error
	WriteFileFromLocal(localPath, remotePath string) error
	AppendFile(content, path string) error
	CopyFile(src, dst string) error
	MoveFile(src, dst string) error
	DeleteFile(path string) error
	CreateDirectory(path string) error
	RemoveDirectory(path string, recursive bool) error
	ListDirectory(path string) Result
	GetFileInfo(path string) FileInfo
	ChangePermissions(path, permissions string) error
	ChangeOwner(path, user, group string) error

	// System operations
	FileExists(path string) bool
	DirectoryExists(path string) bool
	Systemctl(action, service string) bool
	InstallPackage(packageName string) bool

	// Platform detection
	GetDistroInfo() interface{}
	IsUbuntu() bool
	IsDebian() bool
	IsCentOS() bool
	IsRedHat() bool

	// Connection info
	User() string
	Host() string
	Port() int
}

// Result represents command execution result
type Result struct {
	Success   bool
	Output    string
	Error     string
	Stdout    string
	Stderr    string
	ExitCode  int
	Duration  string
	Timestamp string
}

// Result helper methods

// GetError returns the error from the result
func (r *Result) GetError() error {
	if !r.Success && r.Error != "" {
		return fmt.Errorf(r.Error)
	}
	return nil
}

// String returns a string representation of the result
func (r *Result) String() string {
	if r.Success {
		return r.Output
	}
	return r.Error
}

// Lines splits output into lines
func (r *Result) Lines() []string {
	if r.Output == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(r.Output), "\n")
}

// ErrorLines splits error into lines
func (r *Result) ErrorLines() []string {
	if r.Error == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(r.Error), "\n")
}

// Contains checks if output contains the given text
func (r *Result) Contains(text string) bool {
	return strings.Contains(r.Output, text)
}

// ContainsError checks if error contains the given text
func (r *Result) ContainsError(text string) bool {
	return strings.Contains(r.Error, text)
}

// JSON unmarshals output into the provided interface
func (r *Result) JSON(v interface{}) error {
	return json.Unmarshal([]byte(r.Output), v)
}

// FileInfo represents file information
type FileInfo struct {
	Name        string
	Size        int64
	Mode        os.FileMode
	ModTime     time.Time
	IsDir       bool
	Owner       string
	Group       string
	Permissions string
}

// PlatformInfo represents platform information
type PlatformInfo struct {
	OS           string
	Version      string
	Architecture string
	Kernel       string
	IsDocker     bool
	IsVM         bool
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Available uint64
	Cached    uint64
	Buffers   uint64
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Total      uint64
	Used       uint64
	Free       uint64
	Path       string
	MountPath  string
	Filesystem string
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model     string
	Cores     int
	Frequency float64
	Usage     float64
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	ConnectedAt   time.Time
	LastActivity  time.Time
	CommandsRun   int
	BytesSent     int64
	BytesReceived int64
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

// ValidatePlugins validates all registered plugins
func (r *Registry) ValidatePlugins() []error {
	var errors []error

	for name, plugin := range r.plugins {
		if err := plugin.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("plugin %s validation failed: %w", name, err))
		}
	}

	return errors
}

// GetPluginNames returns all plugin names
func (r *Registry) GetPluginNames() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetPluginCount returns the number of registered plugins
func (r *Registry) GetPluginCount() int {
	return len(r.plugins)
}

// Remove removes a plugin by name
func (r *Registry) Remove(name string) bool {
	if _, exists := r.plugins[name]; !exists {
		return false
	}
	delete(r.plugins, name)
	return true
}

// Clear removes all plugins
func (r *Registry) Clear() {
	r.plugins = make(map[string]Plugin)
}

// LoadPlugin loads a plugin by name using the loader and registers it
func (r *Registry) LoadPlugin(name string) error {
	if r.loader == nil {
		return fmt.Errorf("no plugin loader configured")
	}

	plugin, err := r.loader.LoadPlugin(name)
	if err != nil {
		return err
	}

	r.Register(plugin)
	return nil
}

// LoadAll loads all available plugins using the loader
func (r *Registry) LoadAll() error {
	if r.loader == nil {
		return fmt.Errorf("no plugin loader configured")
	}

	plugins, err := r.loader.LoadPlugins()
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		r.Register(plugin)
	}

	return nil
}

// GetRootCommands returns all root commands from registered plugins
func (r *Registry) GetRootCommands() []*cobra.Command {
	var commands []*cobra.Command

	for _, plugin := range r.plugins {
		if cmd := plugin.GetRootCommand(); cmd != nil {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// WithHideOutput returns true to indicate output should be hidden
func WithHideOutput() bool {
	return true
}
