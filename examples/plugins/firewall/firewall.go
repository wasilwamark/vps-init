package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wasilwamark/vps-init-ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// FirewallPlugin implements the plugin interface
type FirewallPlugin struct {
	ssh    ssh.Connection
	config map[string]interface{}
}

// NewPlugin creates a new firewall plugin
func NewPlugin() plugin.Plugin {
	return &FirewallPlugin{}
}

func (p *FirewallPlugin) Name() string {
	return "firewall"
}

func (p *FirewallPlugin) Description() string {
	return "Firewall management using UFW"
}

func (p *FirewallPlugin) Version() string {
	return "1.0.0"
}

func (p *FirewallPlugin) Author() string {
	return "VPS-Init Team"
}

func (p *FirewallPlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *FirewallPlugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install and configure UFW firewall",
			Flags: []plugin.Flag{
				{
					Name:        "default-policy",
					Description: "Default firewall policy (allow/deny)",
					Default:     "deny",
					Type:        plugin.ArgumentTypeString,
				},
				{
					Name:        "enable-logging",
					Description: "Enable firewall logging",
					Default:     true,
					Type:        plugin.ArgumentTypeBool,
				},
			},
			Handler: p.handleInstall,
		},
		{
			Name:        "allow",
			Description: "Allow traffic through firewall",
			Args: []plugin.Argument{
				{
					Name:        "port",
					Description: "Port number or service name",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
				{
					Name:        "protocol",
					Description: "Protocol (tcp/udp)",
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleAllow,
		},
		{
			Name:        "deny",
			Description: "Deny traffic through firewall",
			Args: []plugin.Argument{
				{
					Name:        "port",
					Description: "Port number or service name",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleDeny,
		},
		{
			Name:        "status",
			Description: "Show firewall status",
			Handler:     p.handleStatus,
		},
		{
			Name:        "enable",
			Description: "Enable firewall",
			Handler:     p.handleEnable,
		},
		{
			Name:        "disable",
			Description: "Disable firewall",
			Handler:     p.handleDisable,
		},
	}
}

func (p *FirewallPlugin) GetRootCommand() *cobra.Command {
	// This plugin doesn't add a root command, only subcommands
	return nil
}

func (p *FirewallPlugin) Start(ctx context.Context) error {
	return nil
}

func (p *FirewallPlugin) Stop(ctx context.Context) error {
	return nil
}

func (p *FirewallPlugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{}
}

func (p *FirewallPlugin) Validate() error {
	return nil
}

func (p *FirewallPlugin) Compatibility() plugin.Compatibility {
	return plugin.Compatibility{
		MinVPSInitVersion: "0.0.1",
		GoVersion:         "1.21",
		Platforms:         []string{"linux/amd64", "linux/arm64"},
	}
}

func (p *FirewallPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        p.Name(),
		Description: p.Description(),
		Version:     p.Version(),
		Author:      p.Author(),
		License:     "MIT",
		Tags:        []string{"firewall", "security", "ufw"},
	}
}

// Command handlers
func (p *FirewallPlugin) handleInstall(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn

	fmt.Println("🔥 Installing UFW firewall...")

	// Install UFW
	if !p.ssh.InstallPackage("ufw") {
		return fmt.Errorf("failed to install UFW")
	}

	// Set default policy
	defaultPolicy := "deny"
	if dp, ok := flags["default-policy"].(string); ok && dp != "" {
		defaultPolicy = dp
	}

	result := p.ssh.RunCommand(fmt.Sprintf("ufw default %s", defaultPolicy))
	if !result.Success {
		return fmt.Errorf("failed to set default policy")
	}

	// Enable logging if requested
	enableLogging := true
	if el, ok := flags["enable-logging"].(bool); ok {
		enableLogging = el
	}

	if enableLogging {
		result = p.ssh.RunCommand("ufw logging on")
		if !result.Success {
			return fmt.Errorf("failed to enable logging")
		}
	}

	// Allow SSH by default
	result = p.ssh.RunCommand("ufw allow ssh")
	if !result.Success {
		return fmt.Errorf("failed to allow SSH")
	}

	fmt.Println("✅ UFW firewall installed and configured")
	fmt.Println("📝 Note: Run 'vps-init firewall enable' to activate the firewall")

	return nil
}

func (p *FirewallPlugin) handleAllow(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn

	if len(args) < 1 {
		return fmt.Errorf("port is required")
	}

	port := args[0]
	protocol := ""
	if len(args) > 1 {
		protocol = args[1]
	}

	var cmd string
	if protocol != "" {
		cmd = fmt.Sprintf("ufw allow %s/%s", port, protocol)
	} else {
		cmd = fmt.Sprintf("ufw allow %s", port)
	}

	result := p.ssh.RunCommand(cmd)
	if !result.Success {
		return fmt.Errorf("failed to allow port %s", port)
	}

	fmt.Printf("✅ Allowed traffic on port %s\n", port)
	return nil
}

func (p *FirewallPlugin) handleDeny(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn

	if len(args) < 1 {
		return fmt.Errorf("port is required")
	}

	port := args[0]
	cmd := fmt.Sprintf("ufw deny %s", port)

	result := p.ssh.RunCommand(cmd)
	if !result.Success {
		return fmt.Errorf("failed to deny port %s", port)
	}

	fmt.Printf("✅ Denied traffic on port %s\n", port)
	return nil
}

func (p *FirewallPlugin) handleStatus(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn

	result := p.ssh.RunCommand("ufw status verbose")
	if result.Success {
		fmt.Println(result.Stdout)
	} else {
		return fmt.Errorf("failed to get firewall status")
	}

	return nil
}

func (p *FirewallPlugin) handleEnable(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn

	result := p.ssh.RunCommand("ufw --force enable")
	if !result.Success {
		return fmt.Errorf("failed to enable firewall")
	}

	fmt.Println("✅ Firewall enabled")
	return nil
}

func (p *FirewallPlugin) handleDisable(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn

	result := p.ssh.RunCommand("ufw disable")
	if !result.Success {
		return fmt.Errorf("failed to disable firewall")
	}

	fmt.Println("✅ Firewall disabled")
	return nil
}

func main() {}
