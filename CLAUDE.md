# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VPS-Init is a Go-based CLI tool for server management that works directly over SSH. It features a modular plugin architecture for extensible functionality without requiring agents or configuration management systems.

## Build Commands

```bash
# Standard build
make build

# Build for development with debug symbols
make dev

# Build for multiple platforms
make build-all

# Install locally (copies to ~/bin/ or ~/.local/bin/)
make install-local

# Install globally (requires sudo)
make install

# Clean build artifacts
make clean

# Download dependencies
make deps

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Run tests
make test
```

## Development Commands

### Single Plugin Development

```bash
# Test a specific plugin
go test ./internal/services/redis

# Build and test a plugin
go build -o bin/vps-init ./cmd/vps-init
./bin/vps-init test@localhost redis status

# Run plugin validation
./bin/vps-init plugin validate
```

### Plugin System Testing

```bash
# Test plugin loading
go test ./pkg/plugin

# Test git installer
go test ./pkg/plugin/git_installer
```

### SSH Connection Testing

```bash
# Test SSH connection (use valid server)
./bin/vps-init user@server alias list

# Test with dummy host (for CI)
./bin/vps-init test@localhost system update
```

## Architecture Overview

### Core Components

#### **CLI Layer (`cmd/vps-init/`)**
- `main.go`: Entry point with plugin system initialization
- `plugins_init.go`: Built-in plugin registration

#### **CLI Interface (`internal/cli/`)**
- `cli.go`: Cobra command-line interface setup
- `root.go`: Root command configuration

#### **Plugin System (`pkg/plugin/`)**
- `interface.go`: Core plugin interfaces and types
- `builtin.go`: Built-in plugin registry
- `loader.go`: Filesystem and dynamic plugin loading
- `validation.go`: Plugin validation system
- `git_installer.go`: Git-based plugin installation
- `git_provider.go`: Git VCS operations
- `compatibility.go`: Version compatibility checking

#### **SSH Layer (`internal/ssh/`)**
- `ssh.go`: SSH connection management with sudo support

#### **Plugin Services (`internal/services/`)**
- Each plugin implements the enhanced Plugin interface
- Modular architecture with dependency management

#### **Core Plugins (`internal/core/`)**
- `alias/`: Server alias management
- `plugin-manager/`: Plugin management commands

### Enhanced Plugin Interface

All plugins must implement:

```go
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
```

### Plugin Development

#### **Creating a New Plugin**

1. Create plugin directory: `internal/services/myplugin/`
2. Implement plugin interface (copy existing plugin as template)
3. Add validation, compatibility, and metadata
4. Register in `cmd/vps-init/plugins_init.go`
5. Test with: `./bin/vps-init plugin info myplugin`

#### **Plugin Directory Structure**

```
internal/services/myplugin/
├── plugin.go          # Main plugin implementation
└── plugin.yaml        # Plugin metadata (optional)
```

#### **Git-based Plugin Installation**

```bash
# Install from GitHub
./bin/vps-init plugin install github.com/user/my-plugin

# Install specific version
./bin/vps-init plugin install github.com/user/my-plugin@v1.0.0

# Install from branch
./bin/vps-init plugin install github.com/user/my-plugin --branch main
```

## Key Design Patterns

### **SSH Connection Management**
- Uses standard SSH keys for authentication
- Secure sudo password injection via flags or environment variables
- Automatic connection cleanup

### **Plugin Architecture**
- **Interface-based**: All plugins implement standardized interface
- **Dependency management**: Plugins declare version constraints
- **Validation system**: Comprehensive plugin validation
- **Loading mechanisms**: Built-in, filesystem, and git-based loading

### **Command Execution**
- Plugin commands execute over SSH with context and flags
- Consistent error handling and user feedback
- Flag-based configuration support

### **Security Features**
- Plugin validation before loading
- Trust level system (official, community, untrusted)
- Platform compatibility checking
- Semantic versioning enforcement

## Plugin Registry

### **Builtin Registration**
Plugins are registered in `cmd/vps-init/plugins_init.go`:

```go
plugin.RegisterBuiltin("github.com/wasilwps-init/services/nginx", &nginx.Plugin{})
```

### **Dynamic Loading**
Plugins can be loaded from:
- Local filesystem paths
- Git repositories
- Package imports

### **Plugin Discovery**
```bash
# List all available plugins
./bin/vps-init plugin list

# Show plugin information
./bin/vps-init plugin info nginx

# Validate all plugins
./bin/vps-init plugin validate
```

## Important Implementation Details

### **SSH Connection Usage**
```go
// Always handle SSH connections properly
result := conn.RunCommand("apt update", false)
if !result.Success {
    return fmt.Errorf("command failed: %s", result.Stderr)
}
```

### **Plugin Dependencies**
```go
// WordPress plugin dependencies example
func (p *Plugin) Dependencies() []plugin.Dependency {
    return []plugin.Dependency{
        {
            Name:     "mysql",
            Version:  ">=0.0.1",
            Optional: false,
        },
        {
            Name:     "nginx",
            Version:  ">=0.0.1",
            Optional: false,
        },
    }
}
```

### **Sudo Password Handling**
```go
// Always get sudo pass securely from flags or environment
func getSudoPass(flags map[string]interface{}) string {
    if pass, ok := flags["sudo_password"].(string); ok {
        return pass
    }
    return ""
}
```

## Development Workflow

1. **Code Changes**: Make modifications to existing code
2. **Testing**: `make test` and `./bin/vps-init plugin test`
3. **Building**: `make build` or `make dev`
4. **Installation**: `make install` for local testing
5. **Validation**: `./bin/vps-init plugin validate`

## Testing Strategy

### **Unit Testing**
```bash
# Test specific package
go test ./pkg/plugin

# Test with coverage
go test -cover ./internal/services/redis
```

### **Integration Testing**
```bash
# Test plugin loading
./bin/vps-init plugin list

# Test plugin commands
./bin/vps-init test@localhost redis status
```

### **Validation Testing**
```bash
# Run comprehensive validation
./bin/vps-init plugin validate

# Run strict validation
./bin/vps-init plugin validate --strict
```

## Plugin Management Commands

The enhanced plugin-manager provides:

```bash
# Install from git repository
./bin/vps-init plugin install github.com/user/plugin

# Update plugins
./bin/vps-init plugin update --all
./bin/vps-init plugin update plugin-name

# Remove plugins
./bin/vps-init plugin remove plugin-name --purge

# Search for plugins
./bin/vps-init plugin search database

# Validate plugins
./bin/vps-init plugin validate
```

## Environment Variables

- `SSH_SUDO_PASS_<ALIAS>`: Sudo password for specific alias
- `VPS_INIT_CACHE_DIR`: Custom cache directory for git plugins
- `VPS_INIT_INSTALL_DIR`: Custom installation directory