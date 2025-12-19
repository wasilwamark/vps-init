package pluginmanager

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct {
	config   map[string]interface{}
	registry *plugin.Registry
}

func NewPlugin() plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "plugin-manager"
}

func (p *Plugin) Description() string {
	return "List and manage built-in plugins"
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
			Description: "List all built-in plugins",
			Handler:     p.handleList,
		},
		{
			Name:        "info",
			Description: "Show information about a built-in plugin",
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
			Name:        "validate",
			Description: "Validate built-in plugins",
			Handler:     p.handleValidate,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Built-in plugin management commands",
		Long: `Manage built-in VPS-Init plugins.

Examples:
  vps-init plugin list
  vps-init plugin info nginx
  vps-init plugin validate`,
	}

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all built-in plugins",
		RunE:  p.runList,
	}
	cmd.AddCommand(listCmd)

	// info command
	infoCmd := &cobra.Command{
		Use:   "info [plugin]",
		Short: "Show information about a built-in plugin",
		Long: `Show detailed information about a built-in plugin including its commands and dependencies.

Examples:
  vps-init plugin info nginx
  vps-init plugin info redis`,
		Args: cobra.ExactArgs(1),
		RunE: p.runInfo,
	}
	cmd.AddCommand(infoCmd)

	// validate command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate built-in plugins",
		Long: `Validate all built-in plugins for compatibility and integrity.

Examples:
  vps-init plugin validate
  vps-init plugin validate --strict`,
		RunE: p.runValidate,
	}
	validateCmd.Flags().Bool("strict", false, "Enable strict validation")
	cmd.AddCommand(validateCmd)

	return cmd
}

func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

// Enhanced plugin interface methods
func (p *Plugin) Validate() error {
	// Plugin manager validation logic
	return nil
}

func (p *Plugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{}
}

func (p *Plugin) Compatibility() plugin.Compatibility {
	return plugin.Compatibility{
		MinVPSInitVersion: "1.0.0",
		GoVersion:         "1.19",
		Platforms:         []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"},
		Tags:              []string{"core", "management", "builtin"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "plugin-manager",
		Description: "Built-in plugin management",
		Version:     "1.0.0",
		Author:      "VPS-Init Team",
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-init",
		Tags:        []string{"core", "management", "builtin"},
		Validated:   true,
		TrustLevel:  "official",
		BuildInfo: plugin.BuildInfo{
			GoVersion: "1.21",
		},
	}
}

// Command handlers
func (p *Plugin) handleList(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
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


func (p *Plugin) handleInfo(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
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
			version := ""
			if dep.Version != "" {
				version = " (" + dep.Version + ")"
			}
			fmt.Printf("  - %s%s\n", dep.Name, version)
		}
	}

	return nil
}


// Cobra command runners
func (p *Plugin) runList(cmd *cobra.Command, args []string) error {
	if p.registry == nil {
		p.registry = plugin.GetBuiltinRegistry()
	}
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


func (p *Plugin) runInfo(cmd *cobra.Command, args []string) error {
	if p.registry == nil {
		p.registry = plugin.GetBuiltinRegistry()
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
			version := ""
			if dep.Version != "" {
				version = " (" + dep.Version + ")"
			}
			fmt.Printf("  - %s%s\n", dep.Name, version)
		}
	}

	return nil
}



func (p *Plugin) handleValidate(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	if p.registry == nil {
		p.registry = plugin.GetBuiltinRegistry()
	}

	strict := false
	if flag, ok := flags["strict"].(bool); ok {
		strict = flag
	}

	fmt.Println("✅ Validating installed plugins...")
	if strict {
		fmt.Println("   (strict mode enabled)")
	}

	plugins := p.registry.GetAll()
	if len(plugins) == 0 {
		fmt.Println("No plugins to validate.")
		return nil
	}

	validator := plugin.NewValidator("1.0.0") // Get from build info

	validationErrors := 0
	for _, pl := range plugins {
		fmt.Printf("   Validating '%s'... ", pl.Name())

		errors := validator.ValidatePlugin(pl)
		if len(errors) > 0 {
			fmt.Printf("❌\n")
			validationErrors++
			for _, err := range errors {
				fmt.Printf("      %s\n", err.Error())
			}
		} else {
			fmt.Printf("✅\n")
		}
	}

	if validationErrors > 0 {
		fmt.Printf("\n❌ Validation failed for %d plugins\n", validationErrors)
		return fmt.Errorf("plugin validation failed")
	}

	fmt.Printf("\n✅ All %d plugins validated successfully\n", len(plugins))
	return nil
}


func (p *Plugin) runValidate(cmd *cobra.Command, args []string) error {
	flags := make(map[string]interface{})
	if strict, _ := cmd.Flags().GetBool("strict"); strict {
		flags["strict"] = strict
	}

	return p.handleValidate(context.Background(), nil, args, flags)
}
