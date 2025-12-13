package monitoring

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct {
	service *Service
	config  map[string]interface{}
}

func NewPlugin() plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "monitoring"
}

func (p *Plugin) Description() string {
	return "System monitoring and alerting setup"
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Author() string {
	return "VPS-Init Team"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.service = &Service{}
	return nil
}

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "setup",
			Description: "Set up monitoring and alerting",
			Flags: []plugin.Flag{
				{
					Name:        "uptime-kuma",
					Description: "Install Uptime Kuma",
					Default:     true,
					Type:        plugin.ArgumentTypeBool,
				},
				{
					Name:        "disk-threshold",
					Description: "Disk usage alert threshold (%)",
					Default:     80,
					Type:        plugin.ArgumentTypeInt,
				},
			},
			Handler: p.handleSetup,
		},
		{
			Name:        "dashboard",
			Description: "Show monitoring dashboard URL",
			Handler:     p.handleDashboard,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitoring",
		Short: "System monitoring and alerting",
		Long: `Set up system monitoring and alerting.

Examples:
  vps-init monitoring setup
  vps-init monitoring dashboard`,
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up monitoring and alerting",
		Run:   p.runSetup,
	}
	cmd.AddCommand(setupCmd)

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

func (p *Plugin) handleSetup(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	if !p.service.Setup() {
		return fmt.Errorf("monitoring setup failed")
	}
	return nil
}

func (p *Plugin) handleDashboard(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("📊 Monitoring Dashboard:")
	fmt.Println("  Uptime Kuma: http://your-server-ip:3001")
	fmt.Println("  System Monitor: http://your-server-ip/monitoring.html")
	return nil
}

func (p *Plugin) runSetup(cmd *cobra.Command, args []string) {}