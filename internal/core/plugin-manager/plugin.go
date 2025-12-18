package pluginmanager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	core "github.com/wasilwamark/vps-init-core"
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
			Name:        "install",
			Description: "Install a plugin from git repository",
			Args: []plugin.Argument{
				{
					Name:        "repository",
					Description: "Git repository URL (e.g., github.com/user/plugin)",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleInstall,
		},
		{
			Name:        "update",
			Description: "Update installed plugins",
			Args: []plugin.Argument{
				{
					Name:        "plugin",
					Description: "Plugin name to update (optional, updates all if not specified)",
					Required:    false,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleUpdate,
		},
		{
			Name:        "remove",
			Description: "Remove an installed plugin",
			Args: []plugin.Argument{
				{
					Name:        "plugin",
					Description: "Plugin name to remove",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleRemove,
		},
		{
			Name:        "search",
			Description: "Search for available plugins",
			Args: []plugin.Argument{
				{
					Name:        "query",
					Description: "Search query",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleSearch,
		},
		{
			Name:        "validate",
			Description: "Validate installed plugins",
			Handler:     p.handleValidate,
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
  vps-init plugin install github.com/user/plugin
  vps-init plugin install github.com/user/plugin@v1.2.0
  vps-init plugin update plugin-name
  vps-init plugin info nginx`,
	}

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available plugins",
		RunE:  p.runList,
	}
	cmd.AddCommand(listCmd)

	// install command
	installCmd := &cobra.Command{
		Use:   "install [repository]",
		Short: "Install a plugin from git repository",
		Long: `Install a plugin from a git repository.

Examples:
  vps-init plugin install github.com/user/plugin
  vps-init plugin install github.com/user/plugin@v1.2.0
  vps-init plugin install github.com/user/plugin@main`,
		Args: cobra.ExactArgs(1),
		RunE: p.runInstall,
	}
	installCmd.Flags().String("name", "", "Custom name for the plugin")
	installCmd.Flags().String("version", "", "Specific version to install")
	installCmd.Flags().String("branch", "", "Specific branch to install")
	installCmd.Flags().Bool("force", false, "Force installation even if plugin exists")
	cmd.AddCommand(installCmd)

	// update command
	updateCmd := &cobra.Command{
		Use:   "update [plugin]",
		Short: "Update installed plugins",
		Long: `Update installed plugins to their latest versions.

Examples:
  vps-init plugin update plugin-name
  vps-init plugin update --all`,
		RunE: p.runUpdate,
	}
	updateCmd.Flags().Bool("all", false, "Update all plugins")
	cmd.AddCommand(updateCmd)

	// remove command
	removeCmd := &cobra.Command{
		Use:   "remove [plugin]",
		Short: "Remove an installed plugin",
		Long: `Remove an installed plugin from the system.

Examples:
  vps-init plugin remove plugin-name`,
		Args: cobra.ExactArgs(1),
		RunE: p.runRemove,
	}
	removeCmd.Flags().Bool("purge", false, "Remove all configuration and data")
	cmd.AddCommand(removeCmd)

	// search command
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for available plugins",
		Long: `Search for available plugins in the registry.

Examples:
  vps-init plugin search database
  vps-init plugin search --tag nginx`,
		Args: cobra.ExactArgs(1),
		RunE: p.runSearch,
	}
	searchCmd.Flags().StringSlice("tag", []string{}, "Filter by tags")
	cmd.AddCommand(searchCmd)

	// validate command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate installed plugins",
		Long: `Validate all installed plugins for compatibility and integrity.

Examples:
  vps-init plugin validate
  vps-init plugin validate --strict`,
		RunE: p.runValidate,
	}
	validateCmd.Flags().Bool("strict", false, "Enable strict validation")
	cmd.AddCommand(validateCmd)

	// load command
	loadCmd := &cobra.Command{
		Use:   "load [plugin]",
		Short: "Load a specific plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  p.runLoad,
	}
	cmd.AddCommand(loadCmd)

	// info command
	infoCmd := &cobra.Command{
		Use:   "info [plugin]",
		Short: "Show information about a plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  p.runInfo,
	}
	cmd.AddCommand(infoCmd)

	// reload command
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
		Tags:              []string{"core", "management", "plugin"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "plugin-manager",
		Description: "Plugin management and discovery",
		Version:     "1.0.0",
		Author:      "VPS-Init Team",
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-init",
		Tags:        []string{"core", "management", "plugin"},
		Validated:   true,
		TrustLevel:  "official",
		BuildInfo: plugin.BuildInfo{
			GoVersion: "1.21",
		},
	}
}

// Command handlers
func (p *Plugin) handleList(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
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

func (p *Plugin) handleLoad(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
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

func (p *Plugin) handleInfo(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
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

func (p *Plugin) handleReload(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔄 Reloading plugins...")

	// TODO: Implement plugin reload logic
	fmt.Println("⚠️  Plugin reload not yet implemented")
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

func (p *Plugin) runLoad(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	if err := p.registry.LoadPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
	}

	fmt.Printf("✅ Plugin '%s' loaded successfully\n", pluginName)
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

func (p *Plugin) runReload(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Reloading plugins...")
	fmt.Println("⚠️  Plugin reload not yet implemented")
}

// New command handlers
func (p *Plugin) handleInstall(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("repository URL is required")
	}

	repoURL := args[0]

	// Add https:// prefix if not present and not an SCP-like URL
	if !strings.HasPrefix(repoURL, "https://") &&
	   !strings.HasPrefix(repoURL, "http://") &&
	   !strings.HasPrefix(repoURL, "git@") &&
	   !strings.Contains(repoURL, "://") {
		repoURL = "https://" + repoURL
	}

	fmt.Printf("🔧 Installing plugin from %s\n", repoURL)

	// Extract version from URL if present
	var version string
	if strings.Contains(repoURL, "@") {
		parts := strings.Split(repoURL, "@")
		if len(parts) >= 2 {
			version = parts[1]
		}
	}

	// Get options from flags
	options := plugin.InstallOptions{
		Version: version,
		Force:   false,
	}

	if force, ok := flags["force"].(bool); ok {
		options.Force = force
	}

	if name, ok := flags["name"].(string); ok && name != "" {
		options.Name = name
	}

	if branch, ok := flags["branch"].(string); ok && branch != "" {
		options.Branch = branch
	}

	if ver, ok := flags["version"].(string); ok && ver != "" {
		options.Version = ver
	}

	// Create git installer
	config := plugin.InstallerConfig{
		CacheDir:   "~/.vps-init/cache",
		InstallDir: "~/.vps-init/plugins",
		Timeout:    5 * time.Minute,
	}

	installer := plugin.NewGitInstaller(config)

	// Install plugin
	metadata, err := installer.Install(ctx, repoURL, options)
	if err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	fmt.Printf("✅ Plugin '%s' installed successfully!\n", metadata.Name)
	fmt.Printf("   Version: %s\n", metadata.Version)
	fmt.Printf("   Description: %s\n", metadata.Description)

	return nil
}

func (p *Plugin) handleUpdate(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
	updateAll := false
	if flag, ok := flags["all"].(bool); ok {
		updateAll = flag
	}

	if updateAll {
		fmt.Println("🔄 Updating all plugins...")
		// TODO: Implement update all logic
		fmt.Println("⚠️  Update all functionality not yet implemented")
	} else if len(args) > 0 {
		pluginName := args[0]
		fmt.Printf("🔄 Updating plugin '%s'...\n", pluginName)
		// TODO: Implement update specific plugin logic
		fmt.Printf("⚠️  Update functionality for '%s' not yet implemented\n", pluginName)
	} else {
		return fmt.Errorf("either specify a plugin name or use --all flag")
	}

	return nil
}

func (p *Plugin) handleRemove(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("plugin name is required")
	}

	pluginName := args[0]
	purge := false
	if flag, ok := flags["purge"].(bool); ok {
		purge = flag
	}

	fmt.Printf("🗑️  Removing plugin '%s'...\n", pluginName)
	if purge {
		fmt.Println("   (purging all configuration and data)")
	}

	// TODO: Implement plugin removal logic
	fmt.Printf("⚠️  Plugin removal functionality not yet implemented\n")

	return nil
}

func (p *Plugin) handleSearch(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("search query is required")
	}

	query := args[0]
	var tags []string
	if tagSlice, ok := flags["tag"].([]string); ok {
		tags = tagSlice
	}

	fmt.Printf("🔍 Searching for plugins matching '%s'\n", query)
	if len(tags) > 0 {
		fmt.Printf("   Tags: %v\n", tags)
	}

	// TODO: Implement plugin search logic
	fmt.Println("⚠️  Plugin search functionality not yet implemented")
	fmt.Println("   Available plugins:")
	fmt.Println("   - redis: Redis database server management")
	fmt.Println("   - nginx: Manage Nginx web server")
	fmt.Println("   - docker: Manage Docker Engine & Compose")
	fmt.Println("   - wireguard: Wireguard VPN Server")

	return nil
}

func (p *Plugin) handleValidate(ctx context.Context, conn core.Connection, args []string, flags map[string]interface{}) error {
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

// New Cobra command runners
func (p *Plugin) runInstall(cmd *cobra.Command, args []string) error {
	flags := make(map[string]interface{})

	name, _ := cmd.Flags().GetString("name")
	if name != "" {
		flags["name"] = name
	}

	version, _ := cmd.Flags().GetString("version")
	if version != "" {
		flags["version"] = version
	}

	branch, _ := cmd.Flags().GetString("branch")
	if branch != "" {
		flags["branch"] = branch
	}

	force, _ := cmd.Flags().GetBool("force")
	flags["force"] = force

	return p.handleInstall(context.Background(), nil, args, flags)
}

func (p *Plugin) runUpdate(cmd *cobra.Command, args []string) error {
	flags := make(map[string]interface{})
	if all, _ := cmd.Flags().GetBool("all"); all {
		flags["all"] = all
	}

	return p.handleUpdate(context.Background(), nil, args, flags)
}

func (p *Plugin) runRemove(cmd *cobra.Command, args []string) error {
	flags := make(map[string]interface{})
	if purge, _ := cmd.Flags().GetBool("purge"); purge {
		flags["purge"] = purge
	}

	return p.handleRemove(context.Background(), nil, args, flags)
}

func (p *Plugin) runSearch(cmd *cobra.Command, args []string) error {
	flags := make(map[string]interface{})
	if tags, _ := cmd.Flags().GetStringSlice("tag"); len(tags) > 0 {
		flags["tag"] = tags
	}

	return p.handleSearch(context.Background(), nil, args, flags)
}

func (p *Plugin) runValidate(cmd *cobra.Command, args []string) error {
	flags := make(map[string]interface{})
	if strict, _ := cmd.Flags().GetBool("strict"); strict {
		flags["strict"] = strict
	}

	return p.handleValidate(context.Background(), nil, args, flags)
}
