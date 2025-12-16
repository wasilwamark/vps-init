package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init-ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string {
	return "docker"
}

func (p *Plugin) Description() string {
	return "Manage Docker Engine & Compose"
}

func (p *Plugin) Author() string {
	return "VPS-Init"
}

func (p *Plugin) Version() string {
	return "0.0.1"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	return nil
}

func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}



func (p *Plugin) GetRootCommand() *cobra.Command {
	return nil
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


func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install Docker Engine & Compose",
			Handler:     p.installHandler,
		},
		{
			Name:        "status",
			Description: "Check Docker status",
			Handler:     p.statusHandler,
		},
		{
			Name:        "compose",
			Description: "Run docker compose commands",
			Handler:     p.composeHandler,
		},
		{
			Name:        "ps",
			Description: "List running containers",
			Handler:     p.simpleDockerHandler("ps"),
		},
		{
			Name:        "logs",
			Description: "Stream container logs",
			Handler:     p.logsHandler,
		},
		{
			Name:        "prune",
			Description: "Prune unused docker resources",
			Handler:     p.simpleDockerHandler("system prune -f"),
		},
		{
			Name:        "verify",
			Description: "Verify installation (hello-world)",
			Handler:     p.simpleDockerHandler("run --rm hello-world"),
		},
		// Compose Shortcuts
		{
			Name:        "up",
			Description: "Start services (docker compose up -d)",
			Handler:     p.simpleComposeHandler("up -d"),
		},
		{
			Name:        "down",
			Description: "Stop services (docker compose down)",
			Handler:     p.simpleComposeHandler("down"),
		},
		{
			Name:        "pull",
			Description: "Pull images (docker compose pull)",
			Handler:     p.simpleComposeHandler("pull"),
		},
	}
}

func (p *Plugin) installHandler(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🐳 Installing Docker...")

	// using convenience script
	cmd := "curl -fsSL https://get.docker.com | sh"
	pass := getSudoPass(flags)

	result := conn.RunSudo(cmd, pass); if !result.Success {
		return fmt.Errorf("failed to install docker: %s", result.Stderr)
	}

	// Add user to docker group
	fmt.Println("👤 Adding user to docker group...")
	result = conn.RunSudo("usermod -aG docker $USER", pass); if !result.Success {
		fmt.Printf("⚠️  Failed to add user to docker group: %s\n", result.Stderr)
	} else {
		fmt.Println("✅ User added to docker group (requires re-login to take effect)")
	}

	fmt.Println("✅ Docker installed successfully!")
	return nil
}

func (p *Plugin) statusHandler(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	return conn.RunInteractive("systemctl status docker")
}

func (p *Plugin) composeHandler(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	// Reconstruct command line arguments
	cmd := "docker compose " + strings.Join(args, " ")
	return conn.RunInteractive(cmd)
}

func (p *Plugin) logsHandler(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		// Default to compose logs if no container specified
		fmt.Println("📜 Streaming docker compose logs...")
		return conn.RunInteractive("docker compose logs -f")
	}
	// Specific container logs
	cmd := fmt.Sprintf("docker logs -f %s", args[0])
	return conn.RunInteractive(cmd)
}

func (p *Plugin) simpleDockerHandler(subcmd string) plugin.CommandHandler {
	return func(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
		cmd := fmt.Sprintf("docker %s", subcmd)
		return conn.RunInteractive(cmd)
	}
}

func (p *Plugin) simpleComposeHandler(subcmd string) plugin.CommandHandler {
	return func(ctx context.Context, conn ssh.Connection, args []string, flags map[string]interface{}) error {
		// Allow passing extra args e.g. vps-init alias docker up --build
		argsStr := strings.Join(args, " ")
		cmd := fmt.Sprintf("docker compose %s %s", subcmd, argsStr)
		// Trim space in case argsStr is empty
		cmd = strings.TrimSpace(cmd)
		return conn.RunInteractive(cmd)
	}
}

// Helper
func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}
