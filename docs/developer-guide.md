# Plugin Developer Guide

This document provides a comprehensive guide to understanding the VPS-Init architecture and developing plugins for it.

## 🏗️ Architecture Overview

VPS-Init is built entirely on a **Plugin Architecture**. Even core features like alias management are implemented as plugins.

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
│          Service Plugins                │
│  ┌─────────────┬─────────────┬─────────┐ │
│  │    Nginx    │   Docker    │ Fail2Ban│ │
│  └─────────────┴─────────────┴─────────┘ │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│         Custom Plugins                 │
│  ┌─────────────┬─────────────┬─────────┐ │
│  │   Backup    │   Security  │  Custom │ │
│  └─────────────┴─────────────┴─────────┘ │
└─────────────────────────────────────────┘
```

## 🔌 Plugin Types

1.  **Core Plugins (Built-in)**: Essential functionality (e.g., `alias`, `system`).
2.  **Service Plugins (Built-in)**: Pre-built service management (e.g., `nginx`, `docker`).
3.  **Custom / External Plugins**: User-created plugins loaded dynamically (`.so` files).

## 🚀 How It Works

### Startup Flow

1.  **CLI Initialization**: The CLI starts and initializes the plugin system.
2.  **Discovery**:
    *   **Built-in**: Registers plugins compiled into the binary.
    *   **External**: Scans `~/.vps-init/plugins/` and other paths for `.so` files.
3.  **Registration**: Plugins are registered in the central `Registry`.
4.  **Execution**: Commands are dispatched to the appropriate plugin handler.

### The Plugin Interface

All plugins must implement the `Plugin` interface defined in `pkg/plugin/interface.go`:

```go
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
```

## 💻 Developing a Custom Plugin

You can extend VPS-Init by creating your own plugins.

### 1. Project Structure

Create a new Go project. You will need to import the `plugin` package from `vps-init`.

```go
package main

import (
    "context"
    "github.com/spf13/cobra"
    "github.com/wasilwamark/vps-init/pkg/plugin"
    "github.com/wasilwamark/vps-init-ssh"
)

type MyPlugin struct{}

// Exported symbol for the loader. MUST be named 'NewPlugin'.
func NewPlugin() plugin.Plugin {
    return &MyPlugin{}
}

func (p *MyPlugin) Name() string { return "my-tool" }
func (p *MyPlugin) Description() string { return "My custom VPS tool" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Author() string { return "Me" }
func (p *MyPlugin) Dependencies() []plugin.Dependency { return []plugin.Dependency{} }

func (p *MyPlugin) Initialize(config map[string]interface{}) error { return nil }
func (p *MyPlugin) Validate() error { return nil }
func (p *MyPlugin) Start(ctx context.Context) error { return nil }
func (p *MyPlugin) Stop(ctx context.Context) error { return nil }

func (p *MyPlugin) Compatibility() plugin.Compatibility {
    return plugin.Compatibility{
        MinVPSInitVersion: "0.0.1",
        GoVersion:         "1.21",
        Platforms:         []string{"linux/amd64", "linux/arm64"},
    }
}

func (p *MyPlugin) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name:        p.Name(),
        Description: p.Description(),
        Version:     p.Version(),
        Author:      p.Author(),
        License:     "MIT",
        Tags:        []string{"example"},
    }
}

func (p *MyPlugin) GetCommands() []plugin.Command {
    return []plugin.Command{
        {
            Name:        "do-something",
            Description: "Does something amazing on the VPS",
            Handler: func(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
                return conn.RunInteractive("echo 'Hello from MyPlugin!'")
            },
        },
    }
}

func (p *MyPlugin) GetRootCommand() *cobra.Command { return nil }
```

### 2. Build the Plugin

Go plugins must be built with `-buildmode=plugin`.

**Important**: The plugin must be compiled with the **exact same version** of Go and dependencies as the main `vps-init` binary.

```bash
go build -buildmode=plugin -o my-tool.so main.go
```

### 3. Installation

1.  **Create Plugin Directory**:
    ```bash
    mkdir -p ~/.vps-init/plugins
    ```
2.  **Install**:
    Copy your `.so` file to the plugin directory.
    ```bash
    cp my-tool.so ~/.vps-init/plugins/
    ```

### 4. Verification

Run the following to see your plugin listed:

```bash
vps-init plugin list
```

Usage:
```bash
vps-init <target> my-tool do-something
```

## 📁 Project Directory Structure

```
vps-init/
├── cmd/vps-init/             # CLI entry point
├── internal/
│   ├── cli/                  # CLI coordination
│   ├── services/             # Built-in service plugins (nginx, docker, etc.)
│   └── core/                 # Core plugins (alias, plugin-manager)
├── pkg/plugin/               # Plugin system interfaces & loader
└── docs/                     # Documentation
```
