package system

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// Plugin implements the system upgrade plugin
type Plugin struct {
	ssh    *ssh.Connection
	config map[string]interface{}
}

// NewPlugin creates a new system plugin
func NewPlugin() plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "system"
}

func (p *Plugin) Description() string {
	return "System management and upgrades"
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Author() string {
	return "VPS-Init Team"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "update",
			Description: "Update package lists (apt update)",
			Handler:     p.handleUpdate,
		},
		{
			Name:        "upgrade",
			Description: "Upgrade installed packages (apt upgrade)",
			Handler:     p.handleUpgrade,
		},
		{
			Name:        "full-upgrade",
			Description: "Perform full system upgrade (apt dist-upgrade)",
			Handler:     p.handleFullUpgrade,
		},
		{
			Name:        "autoremove",
			Description: "Remove unused packages",
			Handler:     p.handleAutoremove,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System management commands",
		Long:  "Manage system updates and packages.",
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update package lists (apt update)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("To run on a server, use: vps-init user@host system update")
			return nil
		},
	}
	cmd.AddCommand(updateCmd)

	// Ideally we should add other commands too for consistency in help
	// ...

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

// Command Handlers

func (p *Plugin) handleUpdate(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn
	fmt.Println("🔄 Updating package lists...")

	sudoPass, _ := flags["sudo-password"].(string)
	result := p.ssh.RunSudo("apt-get update", sudoPass)

	if !result.Success {
		return fmt.Errorf("failed to update package lists: %s", result.Stderr)
	}

	fmt.Println("✅ Package lists updated")
	return nil
}

func (p *Plugin) handleUpgrade(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn
	fmt.Println("⬆️  Upgrading packages...")

	sudoPass, _ := flags["sudo-password"].(string)
	// DEBIAN_FRONTEND=noninteractive to avoid prompts
	result := p.ssh.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get upgrade -y", sudoPass)
	if !result.Success {
		return fmt.Errorf("failed to upgrade packages: %s", result.Stderr)
	}

	fmt.Println("✅ Packages upgraded")
	return nil
}

func (p *Plugin) handleFullUpgrade(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn
	fmt.Println("🚀 Performing full system upgrade...")

	sudoPass, _ := flags["sudo-password"].(string)
	result := p.ssh.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get dist-upgrade -y", sudoPass)
	if !result.Success {
		return fmt.Errorf("failed to perform full upgrade: %s", result.Stderr)
	}

	fmt.Println("✅ System fully upgraded")
	return nil
}

func (p *Plugin) handleAutoremove(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.ssh = conn
	fmt.Println("🧹 Removing unused packages...")

	sudoPass, _ := flags["sudo-password"].(string)
	result := p.ssh.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get autoremove -y", sudoPass)
	if !result.Success {
		return fmt.Errorf("failed to autoremove: %s", result.Stderr)
	}

	fmt.Println("✅ Unused packages removed")
	return nil
}
