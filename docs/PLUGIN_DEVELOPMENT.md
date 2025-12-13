# VPS-Init Plugin Development Guide

VPS-Init uses a powerful plugin architecture that allows you to extend its functionality with custom commands and services. This guide shows you how to create and use plugins.

## 🏗️ Plugin Architecture

### What is a Plugin?

A plugin is a Go module that implements the `Plugin` interface and can be dynamically loaded into VPS-Init to provide additional functionality.

### Types of Plugins

1. **Built-in Plugins**: Compiled into VPS-Init (nginx, docker, monitoring)
2. **Shared Object Plugins**: Compiled as `.so` files and loaded dynamically
3. **Go Module Plugins**: Imported as Go packages during build time

## 📋 Plugin Interface

Every plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    // Metadata
    Name() string
    Description() string
    Version() string
    Author() string

    // Initialization
    Initialize(config map[string]interface{}) error

    // Commands
    GetCommands() []Command
    GetRootCommand() *cobra.Command

    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error

    // Dependencies
    Dependencies() []string
}
```

## 🚀 Creating Your First Plugin

### Step 1: Create Plugin Directory

```bash
mkdir my-plugin
cd my-plugin
```

### Step 2: Create Plugin Code

Create `my-plugin.go`:

```go
package main

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/wasilwamark/vps-init/internal/ssh"
    "github.com/wasilwamark/vps-init/pkg/plugin"
)

// MyPlugin implements the Plugin interface
type MyPlugin struct {
    ssh    *ssh.Connection
    config map[string]interface{}
}

// NewPlugin creates a new instance of your plugin
func NewPlugin() plugin.Plugin {
    return &MyPlugin{}
}

// Plugin metadata
func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Description() string {
    return "My awesome VPS-Init plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Author() string {
    return "Your Name"
}

// Initialize is called when the plugin is loaded
func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    p.config = config
    return nil
}

// GetCommands returns the commands this plugin provides
func (p *MyPlugin) GetCommands() []plugin.Command {
    return []plugin.Command{
        {
            Name:        "hello",
            Description: "Say hello to the world",
            Handler:     p.handleHello,
        },
        {
            Name:        "install",
            Description: "Install something cool",
            Handler:     p.handleInstall,
        },
    }
}

// GetRootCommand returns the cobra command for this plugin
func (p *MyPlugin) GetRootCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "my-plugin",
        Short: "My awesome plugin",
        Long:  "This plugin does amazing things",
    }

    helloCmd := &cobra.Command{
        Use:   "hello",
        Short: "Say hello",
        Run:   p.runHello,
    }
    cmd.AddCommand(helloCmd)

    return cmd
}

// Lifecycle methods
func (p *MyPlugin) Start(ctx context.Context) error {
    fmt.Println("Plugin started")
    return nil
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    fmt.Println("Plugin stopped")
    return nil
}

// Dependencies
func (p *MyPlugin) Dependencies() []string {
    return []string{} // or return []string{"nginx"} if it depends on nginx
}

// Command handlers
func (p *MyPlugin) handleHello(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
    p.ssh = conn

    result := p.ssh.RunCommand("echo 'Hello from my plugin!'", false)
    if result.Success {
        fmt.Println("✅ Command executed successfully")
        fmt.Println(result.Stdout)
    }

    return nil
}

func (p *MyPlugin) handleInstall(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
    p.ssh = conn

    // Your installation logic here
    if !p.ssh.InstallPackage("your-package") {
        return fmt.Errorf("failed to install package")
    }

    fmt.Println("✅ Installation completed")
    return nil
}

// Cobra command runners (for CLI integration)
func (p *MyPlugin) runHello(cmd *cobra.Command, args []string) {
    // This will be called when using: vps-init my-plugin hello
}
```

### Step 3: Build as Shared Object

Create `build.sh`:

```bash
#!/bin/bash
echo "🔨 Building my-plugin..."

# Build as plugin
go build -buildmode=plugin -o my-plugin.so my-plugin.go

if [ $? -eq 0 ]; then
    echo "✅ Plugin built successfully: my-plugin.so"
    echo ""
    echo "To use:"
    echo "1. Copy my-plugin.so to ~/.vps-init/plugins/"
    echo "2. Add to ~/.vps-init/plugins.yaml"
else
    echo "❌ Failed to build plugin"
    exit 1
fi
```

```bash
chmod +x build.sh
./build.sh
```

### Step 4: Configure Plugin

Edit `~/.vps-init/plugins.yaml`:

```yaml
plugins:
  my-plugin:
    enabled: true
    path: "~/.vps-init/plugins/my-plugin.so"
    config:
      some_option: "value"

paths:
  - "~/.vps-init/plugins"
```

### Step 5: Use Your Plugin

```bash
# List all plugins
vps-init plugin list

# Use your plugin
vps-init user@server my-plugin hello
vps-init user@server my-plugin install

# Or use as subcommand
vps-init my-plugin hello
```

## 🔧 Advanced Plugin Features

### Plugin Configuration

Access plugin configuration in your handlers:

```go
func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    p.config = config

    // Access configuration
    if someOption, ok := config["some_option"].(string); ok {
        fmt.Println("Option:", someOption)
    }

    return nil
}
```

### Command Arguments and Flags

Define complex commands:

```go
{
    Name:        "deploy",
    Description: "Deploy an application",
    Args: []plugin.Argument{
        {
            Name:        "app-name",
            Description: "Name of the application",
            Required:    true,
            Type:        plugin.ArgumentTypeString,
        },
        {
            Name:        "replicas",
            Description: "Number of replicas",
            Type:        plugin.ArgumentTypeInt,
        },
    },
    Flags: []plugin.Flag{
        {
            Name:        "port",
            Shorthand:   "p",
            Description: "Port to deploy on",
            Default:     8080,
            Type:        plugin.ArgumentTypeInt,
        },
        {
            Name:        "env",
            Description: "Environment (dev/staging/prod)",
            Default:     "dev",
            Type:        plugin.ArgumentTypeString,
        },
    },
    Handler: p.handleDeploy,
}
```

### Plugin Dependencies

Declare dependencies on other plugins:

```go
func (p *MyPlugin) Dependencies() []string {
    return []string{"nginx", "docker"}
}
```

VPS-Init will ensure these are available before running your plugin.

## 📦 Plugin Distribution

### Option 1: Shared Object (.so)

```bash
# Build
go build -buildmode=plugin -o my-plugin.so my-plugin.go

# Distribute the .so file
```

### Option 2: Go Module

Create a separate repository and import in `go.mod`:

```go
// In your main project
import "github.com/wasilwamark/vps-init-plugins/myplugin"
```

### Option 3: Built-in

Contribute to the main VPS-Init repository as a built-in plugin.

## 🌟 Example: Custom Firewall Plugin

See `examples/plugins/firewall/` for a complete example plugin that manages UFW firewall rules.

## 📚 Best Practices

1. **Error Handling**: Always return meaningful errors
2. **Logging**: Use structured logging for debugging
3. **Idempotency**: Make commands safe to run multiple times
4. **Configuration**: Support configuration via plugin config
5. **Testing**: Write tests for your plugin logic
6. **Documentation**: Include clear command descriptions

## 🔍 Debugging Plugins

### Enable Debug Mode

```bash
# Set environment variable
export VPS_INIT_DEBUG=1

# Run with debug output
vps-init user@server my-plugin hello
```

### Check Plugin Loading

```bash
# List loaded plugins
vps-init plugin list

# Get plugin info
vps-init plugin info my-plugin
```

### Test Plugin Manually

```go
# Create a test file
package main

import (
    "testing"
    "github.com/wasilwamark/vps-init/pkg/plugin"
)

func TestMyPlugin(t *testing.T) {
    p := NewPlugin()

    if p.Name() != "my-plugin" {
        t.Errorf("Expected name 'my-plugin', got '%s'", p.Name())
    }

    // Test initialization
    err := p.Initialize(map[string]interface{}{"test": true})
    if err != nil {
        t.Errorf("Initialization failed: %v", err)
    }
}
```

## 🤝 Contributing

Share your plugins with the community:

1. **Create a repository**: `github.com/wasilwamark/vps-init-plugins/my-plugin`
2. **Add documentation**: Include README with usage examples
3. **Tag releases**: Follow semantic versioning
4. **Submit to list**: Add to VPS-Init plugins registry

## 🆘 Need Help?

- Check the example plugins in `examples/plugins/`
- Review existing plugins in `internal/services/`
- Open an issue on GitHub for questions

---

Happy coding! 🚀