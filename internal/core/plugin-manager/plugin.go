package pluginmanager

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct {
	config  map[string]interface{}
	registry *plugin.Registry
}

func NewPlugin() plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "plugin-manager"
}

func (p *Plugin) Description() string {
	return "Plugin management and discovery"
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Author() string {
	return "VPS-Init Team"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.registry = plugin.GetBuiltinRegistry()
	return nil
}

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "list",
			Description: "List all available plugins",
			Handler:     p.handleList,
		},
		{
			Name:        "load",
			Description: "Load a specific plugin",
			Args: []plugin.Argument{
				{
					Name:        "name",
					Description: "Plugin name to load",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleLoad,
		},
		{
			Name:        "info",
			Description: "Show information about a plugin",
			Args: []plugin.Argument{
				{
					Name:        "name",
					Description: "Plugin name",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleInfo,
		},
		{
			Name:        "reload",
			Description: "Reload all plugins",
			Handler:     p.handleReload,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin management commands",
		Long: `Manage VPS-Init plugins.

Examples:
  vps-init plugin list
  vps-init plugin load nginx
  vps-init plugin info nginx`,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available plugins",
		RunE:  p.runList,
	}
	cmd.AddCommand(listCmd)

	loadCmd := &cobra.Command{
		Use:   "load [plugin]",
		Short: "Load a specific plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  p.runLoad,
	}
	cmd.AddCommand(loadCmd)

	infoCmd := &cobra.Command{
		Use:   "info [plugin]",
		Short: "Show information about a plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  p.runInfo,
	}
	cmd.AddCommand(infoCmd)

	reloadCmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload all plugins",
		Run:   p.runReload,
	}
	cmd.AddCommand(reloadCmd)

	return cmd
}

func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

func (p *Plugin) Dependencies() []string {
	return []string{}
}

// Command handlers
func (p *Plugin) handleList(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	plugins := p.registry.GetAll()

	if len(plugins) == 0 {
		fmt.Println("No plugins loaded.")
		return nil
	}

	fmt.Println("Available Plugins:")
	for _, pl := range plugins {
		fmt.Printf("  %s (%s) - %s\n", pl.Name(), pl.Version(), pl.Description())
	}

	return nil
}

func (p *Plugin) handleLoad(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("plugin name is required")
	}

	pluginName := args[0]

	if err := p.registry.LoadPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
	}

	fmt.Printf("✅ Plugin '%s' loaded successfully\n", pluginName)
	return nil
}

func (p *Plugin) handleInfo(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("plugin name is required")
	}

	pluginName := args[0]

	pl, exists := p.registry.Get(pluginName)
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	fmt.Printf("Plugin: %s\n", pl.Name())
	fmt.Printf("Version: %s\n", pl.Version())
	fmt.Printf("Description: %s\n", pl.Description())
	fmt.Printf("Author: %s\n", pl.Author())

	// List commands
	commands := pl.GetCommands()
	if len(commands) > 0 {
		fmt.Printf("\nCommands:\n")
		for _, cmd := range commands {
			fmt.Printf("  %s - %s\n", cmd.Name, cmd.Description)
		}
	}

	// List dependencies
	deps := pl.Dependencies()
	if len(deps) > 0 {
		fmt.Printf("\nDependencies:\n")
		for _, dep := range deps {
			fmt.Printf("  - %s\n", dep)
		}
	}

	return nil
}

func (p *Plugin) handleReload(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔄 Reloading plugins...")

	// TODO: Implement plugin reload logic
	fmt.Println("⚠️  Plugin reload not yet implemented")
	return nil
}

// Cobra command runners
func (p *Plugin) runList(cmd *cobra.Command, args []string) error {
	plugins := p.registry.GetAll()

	if len(plugins) == 0 {
		fmt.Println("No plugins loaded.")
		return nil
	}

	fmt.Println("Available Plugins:")
	for _, pl := range plugins {
		fmt.Printf("  %s (%s) - %s\n", pl.Name(), pl.Version(), pl.Description())
	}

	return nil
}

func (p *Plugin) runLoad(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	if err := p.registry.LoadPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
	}

	fmt.Printf("✅ Plugin '%s' loaded successfully\n", pluginName)
	return nil
}

func (p *Plugin) runInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	pl, exists := p.registry.Get(pluginName)
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	fmt.Printf("Plugin: %s\n", pl.Name())
	fmt.Printf("Version: %s\n", pl.Version())
	fmt.Printf("Description: %s\n", pl.Description())
	fmt.Printf("Author: %s\n", pl.Author())

	// List commands
	commands := pl.GetCommands()
	if len(commands) > 0 {
		fmt.Printf("\nCommands:\n")
		for _, cmd := range commands {
			fmt.Printf("  %s - %s\n", cmd.Name, cmd.Description)
		}
	}

	// List dependencies
	deps := pl.Dependencies()
	if len(deps) > 0 {
		fmt.Printf("\nDependencies:\n")
		for _, dep := range deps {
			fmt.Printf("  - %s\n", dep)
		}
	}

	return nil
}

func (p *Plugin) runReload(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Reloading plugins...")
	fmt.Println("⚠️  Plugin reload not yet implemented")
}