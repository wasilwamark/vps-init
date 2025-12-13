package nginx

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// Plugin implements the plugin interface for Nginx
type Plugin struct {
	service *Service
	config  map[string]interface{}
}

// NewPlugin creates a new Nginx plugin
func NewPlugin() plugin.Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "nginx"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Nginx web server management"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Author returns the plugin author
func (p *Plugin) Author() string {
	return "VPS-Init Team"
}

// Initialize initializes the plugin with configuration
func (p *Plugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.service = &Service{}
	return nil
}

// GetCommands returns all commands provided by this plugin
func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install Nginx web server",
			Args:        []plugin.Argument{},
			Flags: []plugin.Flag{
				{
					Name:        "version",
					Shorthand:   "v",
					Description: "Nginx version to install",
					Default:     "latest",
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleInstall,
		},
		{
			Name:        "install-ssl",
			Description: "Install Nginx with Let's Encrypt SSL",
			Args: []plugin.Argument{
				{
					Name:        "domain",
					Description: "Domain name for SSL certificate",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Flags: []plugin.Flag{
				{
					Name:        "email",
					Shorthand:   "e",
					Description: "Email for Let's Encrypt",
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleInstallSSL,
		},
		{
			Name:        "create-site",
			Description: "Create a new Nginx site configuration",
			Args: []plugin.Argument{
				{
					Name:        "domain",
					Description: "Domain name for the site",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Flags: []plugin.Flag{
				{
					Name:        "proxy",
					Description: "Proxy pass target",
					Default:     "http://localhost:3000",
					Type:        plugin.ArgumentTypeString,
				},
				{
					Name:        "ssl",
					Description: "Enable SSL",
					Default:     false,
					Type:        plugin.ArgumentTypeBool,
				},
			},
			Handler: p.handleCreateSite,
		},
		{
			Name:        "reload",
			Description: "Reload Nginx configuration",
			Handler:     p.handleReload,
		},
		{
			Name:        "status",
			Description: "Check Nginx status",
			Handler:     p.handleStatus,
		},
	}
}

// GetRootCommand returns the root cobra command for this plugin
func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nginx",
		Short: "Nginx web server management",
		Long: `Manage Nginx web server configuration and SSL certificates.

Examples:
  vps-init nginx install
  vps-init nginx install-ssl api.tiza.africa
  vps-init nginx create-site myapp.com --proxy=http://localhost:8080`,
	}

	// Add subcommands
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install Nginx web server",
		Run:   p.runInstall,
	}
	installCmd.Flags().StringP("version", "v", "latest", "Nginx version to install")
	cmd.AddCommand(installCmd)

	installSSLCmd := &cobra.Command{
		Use:   "install-ssl [domain]",
		Short: "Install Nginx with Let's Encrypt SSL",
		Args:  cobra.ExactArgs(1),
		Run:   p.runInstallSSL,
	}
	installSSLCmd.Flags().StringP("email", "e", "", "Email for Let's Encrypt")
	cmd.AddCommand(installSSLCmd)

	createSiteCmd := &cobra.Command{
		Use:   "create-site [domain]",
		Short: "Create a new Nginx site configuration",
		Args:  cobra.ExactArgs(1),
		Run:   p.runCreateSite,
	}
	createSiteCmd.Flags().String("proxy", "http://localhost:3000", "Proxy pass target")
	createSiteCmd.Flags().Bool("ssl", false, "Enable SSL")
	cmd.AddCommand(createSiteCmd)

	reloadCmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload Nginx configuration",
		Run:   p.runReload,
	}
	cmd.AddCommand(reloadCmd)

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check Nginx status",
		Run:   p.runStatus,
	}
	cmd.AddCommand(statusCmd)

	return cmd
}

// Start starts the plugin
func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

// Stop stops the plugin
func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

// Dependencies returns the plugin dependencies
func (p *Plugin) Dependencies() []string {
	return []string{}
}

// Command handlers
func (p *Plugin) handleInstall(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	if !p.service.Install() {
		return fmt.Errorf("nginx installation failed")
	}
	return nil
}

func (p *Plugin) handleInstallSSL(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("domain is required")
	}
	p.service = New(conn)
	if !p.service.InstallWithSSL(args[0]) {
		return fmt.Errorf("nginx SSL installation failed")
	}
	return nil
}

func (p *Plugin) handleCreateSite(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("domain is required")
	}
	p.service = New(conn)
	if !p.service.CreateSite(args[0]) {
		return fmt.Errorf("nginx site creation failed")
	}
	return nil
}

func (p *Plugin) handleReload(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	return p.service.Reload()
}

func (p *Plugin) handleStatus(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	return p.service.Status()
}

// Cobra command runners
func (p *Plugin) runInstall(cmd *cobra.Command, args []string) {
	// This will be called by the CLI framework
	// The actual execution will be handled through the plugin system
}

func (p *Plugin) runInstallSSL(cmd *cobra.Command, args []string) {
	// This will be called by the CLI framework
}

func (p *Plugin) runCreateSite(cmd *cobra.Command, args []string) {
	// This will be called by the CLI framework
}

func (p *Plugin) runReload(cmd *cobra.Command, args []string) {
	// This will be called by the CLI framework
}

func (p *Plugin) runStatus(cmd *cobra.Command, args []string) {
	// This will be called by the CLI framework
}