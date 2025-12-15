package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// Plugin implements the system upgrade plugin
type Plugin struct {
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
	// Enhanced plugin interface methods
func (p *Plugin) Validate() error {
	// TODO: Add plugin-specific validation logic
	return nil
}

func (p *Plugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{
		// TODO: Add plugin dependencies with version constraints
	}
}

func (p *Plugin) Compatibility() plugin.Compatibility {
	return plugin.Compatibility{
		MinVPSInitVersion: "1.0.0",
		GoVersion:         "1.19",
		Platforms:         []string{"linux/amd64", "linux/arm64"},
		Tags:              []string{"TODO", "add", "relevant", "tags"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        p.Name(),
		Description: p.Description(),
		Version:     p.Version(),
		Author:      p.Author(),
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-init-plugins/" + p.Name(),
		Tags:        []string{"TODO", "add", "tags"},
		Validated:   true,
		TrustLevel:  "official",
		BuildInfo: plugin.BuildInfo{
			GoVersion: "1.21",
		},
	}
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
		{
			Name:        "shell",
			Description: "Open interactive shell on server",
			Handler:     p.handleShell,
		},
		{
			Name:        "install",
			Description: "Install packages (apt install)",
			Handler:     p.handleInstall,
		},
		{
			Name:        "uninstall",
			Description: "Uninstall packages (apt remove)",
			Handler:     p.handleUninstall,
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



// Command Handlers

// Helper for sudo errors
func (p *Plugin) checkSudoResult(result *ssh.CommandResult, flags map[string]interface{}) error {
	if result.Success {
		return nil
	}

	// Check if sudo password was provided
	sudoPass, _ := flags["sudo-password"].(string)

	errMsg := fmt.Sprintf("failed to execute command: %s", result.Stderr)

	// If it looks like a sudo/permission error
	if strings.Contains(result.Stderr, "sudo") || strings.Contains(result.Stderr, "permission") || strings.Contains(result.Stderr, "password") {
		if sudoPass == "" {
			errMsg += "\n\n❌ Root privileges required.\n"
			errMsg += "Resolution Tips:\n"
			errMsg += "1. Set environment variable: export SSH_SUDO_PWD_<ALIAS>='your-password'\n"
			errMsg += "2. OR Update alias with password: vps-init alias add <name> <user@host> --sudo-password 'pass' (will update existing)\n"
		} else {
			errMsg += "\n\n❌ Sudo authentication failed. Check your password."
		}
	}

	return fmt.Errorf(errMsg)
}

func (p *Plugin) handleUpdate(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔄 Updating package lists...")

	sudoPass, _ := flags["sudo-password"].(string)
	result := conn.RunSudo("apt-get update", sudoPass)

	if err := p.checkSudoResult(result, flags); err != nil {
		return err
	}

	fmt.Println("✅ Package lists updated")
	return nil
}

func (p *Plugin) handleUpgrade(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("⬆️  Upgrading packages...")

	sudoPass, _ := flags["sudo-password"].(string)
	// DEBIAN_FRONTEND=noninteractive to avoid prompts
	result := conn.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get upgrade -y", sudoPass)
	if err := p.checkSudoResult(result, flags); err != nil {
		return err
	}

	fmt.Println("✅ Packages upgraded")
	return nil
}

func (p *Plugin) handleFullUpgrade(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🚀 Performing full system upgrade...")

	sudoPass, _ := flags["sudo-password"].(string)
	result := conn.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get dist-upgrade -y", sudoPass)
	if err := p.checkSudoResult(result, flags); err != nil {
		return err
	}

	fmt.Println("✅ System fully upgraded")
	return nil
}

func (p *Plugin) handleAutoremove(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🧹 Removing unused packages...")

	sudoPass, _ := flags["sudo-password"].(string)
	result := conn.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get autoremove -y", sudoPass)
	if err := p.checkSudoResult(result, flags); err != nil {
		return err
	}

	fmt.Println("✅ Unused packages removed")
	return nil
}

func (p *Plugin) handleShell(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Printf("🔌 Connecting to %s@%s...\n", conn.User, conn.Host)
	return conn.Shell()
}

func (p *Plugin) handleInstall(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: install <package1> [package2...]")
	}

	packages := strings.Join(args, " ")
	fmt.Printf("📦 Installing: %s...\n", packages)

	sudoPass, _ := flags["sudo-password"].(string)
	// -y to assume yes
	cmd := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", packages)
	result := conn.RunSudo(cmd, sudoPass)

	if err := p.checkSudoResult(result, flags); err != nil {
		return err
	}

	fmt.Println("✅ Installation complete")
	return nil
}

func (p *Plugin) handleUninstall(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: uninstall <package1> [package2...]")
	}

	packages := strings.Join(args, " ")
	fmt.Printf("🗑️  Uninstalling: %s...\n", packages)

	sudoPass, _ := flags["sudo-password"].(string)
	// -y to assume yes
	cmd := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get remove -y %s", packages)
	result := conn.RunSudo(cmd, sudoPass)

	if err := p.checkSudoResult(result, flags); err != nil {
		return err
	}

	fmt.Println("✅ Uninstallation complete")
	return nil
}
