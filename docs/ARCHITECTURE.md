# VPS-Init Architecture

## Pure Plugin-Based Design

VPS-Init is built entirely on a plugin architecture. **Even core features are plugins** 

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────┐
│              VPS-Init CLI                │
│         (No hardcoded commands)         │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│         Plugin Registry                  │
│  - Discovers plugins                   │
│  - Loads plugins dynamically          │
│  - Manages plugin lifecycle           │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│            Core Plugins                 │
│  ┌─────────────┬─────────────┬─────────┐ │
│  │    Alias    │Plugin-Manager│   Help  │ │
│  │  Management │   System     │ System  │ │
│  └─────────────┴─────────────┴─────────┘ │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│          Service Plugins                │
│  ┌─────────────┬─────────────┬─────────┐ │
│  │    Nginx    │   Docker    │Monitoring│ │
│  │             │             │         │ │
│  └─────────────┴─────────────┴─────────┘ │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│         Custom Plugins                 │
│  ┌─────────────┬─────────────┬─────────┐ │
│  │   Firewall  │   Backup     │   SSL   │ │
│  │             │             │         │ │
│  └─────────────┴─────────────┴─────────┘ │
└─────────────────────────────────────────┘
```

## 🔌 Plugin Types

### 1. Core Plugins (Built-in)
Essential functionality that ships with VPS-Init:

- **Alias Plugin**: Server alias management
  ```bash
  vps-init alias add myserver user@host.com
  vps-init alias list
  vps-init alias remove myserver
  ```

- **Plugin Manager Plugin**: Plugin system management
  ```bash
  vps-init plugin list
  vps-init plugin info nginx
  vps-init plugin load custom-plugin
  ```

### 2. Service Plugins (Built-in)
Pre-built service management:

- **Nginx Plugin**: Web server management
- **Docker Plugin**: Container platform
- **Monitoring Plugin**: System monitoring

### 3. Custom Plugins
User-created plugins for any functionality:

- Can be compiled as `.so` files
- Can be Go modules
- Can be distributed independently

## 🚀 How It Works

### 1. Startup Flow
```go
main.go
├── InitPluginSystem()
│   ├── Load plugin configuration
│   ├── Initialize plugin loader
│   ├── Discover plugins
│   └── Register all plugins
├── Add plugin commands to CLI
└── Execute CLI
```

### 2. Command Execution
```
User Command: vps-init user@host nginx install
                     │
                     ▼
        ┌─────────────────────────┐
        │    Plugin Registry       │
        │  - Find 'nginx' plugin  │
        └─────────────┬───────────┘
                      │
                      ▼
            ┌─────────────────┐
            │  Nginx Plugin   │
            │  - Execute 'install' │
            └─────────────────┘
```

### 3. Plugin Interface
```go
type Plugin interface {
    // Metadata
    Name() string
    Description() string
    Version() string
    Author() string

    // Commands
    GetCommands() []Command
    GetRootCommand() *cobra.Command

    // Lifecycle
    Initialize(config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error

    // Dependencies
    Dependencies() []string
}
```

## 📁 Directory Structure

```
vps-init/
├── cmd/vps-init/             # CLI entry point
├── internal/
│   ├── cli/                  # CLI coordination
│   ├── core/                 # Core plugins
│   │   ├── alias/           # Alias management plugin
│   │   └── plugin-manager/  # Plugin system plugin
│   ├── services/            # Service plugins
│   │   ├── nginx/           # Nginx plugin
│   │   ├── docker/          # Docker plugin
│   │   └── monitoring/      # Monitoring plugin
│   └── config/              # Configuration management
├── pkg/plugin/              # Plugin system
│   ├── interface.go         # Plugin interface
│   ├── loader.go           # Plugin loading
│   └── builtin.go          # Built-in registry
├── examples/plugins/        # Example custom plugins
└── docs/                    # Documentation
```

## 🔌 Creating Custom Plugins

### 1. Create Plugin File
```go
package main

import "github.com/wasilwamark/vps-init/pkg/plugin"

// MyPlugin implements the Plugin interface
type MyPlugin struct{}

func NewPlugin() plugin.Plugin {
    return &MyPlugin{}
}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Description() string { return "My custom plugin" }
// ... implement other interface methods
```

### 2. Build as Shared Object
```bash
go build -buildmode=plugin -o my-plugin.so my-plugin.go
```

### 3. Configure Plugin
```yaml
# ~/.vps-init/plugins.yaml
plugins:
  my-plugin:
    enabled: true
    path: "~/.vps-init/plugins/my-plugin.so"
```

### 4. Use Your Plugin
```bash
vps-init user@host my-plugin command
```

## 🎯 Benefits

### 1. Extensibility
- Add any functionality without modifying core
- Plugins can be distributed independently
- Community can build and share plugins

### 2. Maintainability
- Core stays simple and clean
- Each plugin is self-contained
- Easy to test and debug individual plugins

### 3. Flexibility
- Users can pick and choose plugins
- Different environments can have different plugins
- Plugin versions can be managed independently

### 4. Performance
- Only load needed plugins
- Lazy loading possible
- Plugins can be unloaded/reloaded

## 🛠️ Plugin Discovery

VPS-Init discovers plugins from multiple sources:

1. **Built-in**: Compiled into VPS-Init
2. **Shared Objects**: `.so` files in plugin paths
3. **Go Modules**: Imported during build time
4. **Remote**: Downloaded from URLs (future feature)

### Plugin Search Paths
```
~/.vps-init/plugins/
./plugins/
/usr/local/lib/vps-init/plugins/
```

## 🔒 Security

- Plugins run with same permissions as VPS-Init
- Plugin configuration is validated
- Plugin loading is sandboxed (future improvement)

## 📚 Examples

See `examples/plugins/` for complete plugin examples:
- **Firewall Plugin**: UFW firewall management
- **Backup Plugin**: Automated backups
- **SSL Plugin**: Certificate management

---

This pure plugin architecture makes VPS-Init incredibly flexible and extensible while keeping the core minimal and focused.