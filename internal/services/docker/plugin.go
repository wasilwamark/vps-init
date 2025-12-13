package docker

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
	return "docker"
}

func (p *Plugin) Description() string {
	return "Docker container platform management"
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
			Name:        "install",
			Description: "Install Docker and Docker Compose",
			Flags: []plugin.Flag{
				{
					Name:        "version",
					Description: "Docker version to install",
					Default:     "latest",
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleInstall,
		},
		{
			Name:        "deploy",
			Description: "Deploy with docker-compose",
			Args: []plugin.Argument{
				{
					Name:        "compose-file",
					Description: "Path to docker-compose.yml",
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleDeploy,
		},
		{
			Name:        "status",
			Description: "Check Docker status",
			Handler:     p.handleStatus,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker",
		Short: "Docker container platform management",
		Long: `Manage Docker container platform.

Examples:
  vps-init docker install
  vps-init docker deploy`,
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install Docker and Docker Compose",
		Run:   p.runInstall,
	}
	cmd.AddCommand(installCmd)

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

func (p *Plugin) handleInstall(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	if !p.service.Install() {
		return fmt.Errorf("docker installation failed")
	}
	return nil
}

func (p *Plugin) handleDeploy(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	composeFile := "docker-compose.yml"
	if len(args) > 0 {
		composeFile = args[0]
	}
	if !p.service.Deploy(composeFile) {
		return fmt.Errorf("docker deploy failed")
	}
	return nil
}

func (p *Plugin) handleStatus(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	p.service = New(conn)
	return p.service.Status()
}

func (p *Plugin) runInstall(cmd *cobra.Command, args []string) {}

// Add missing Service methods
func (s *Service) Status() error {
	fmt.Println("📊 Checking Docker status...")

	result := s.ssh.RunCommand("systemctl status docker --no-pager", false)
	if result.Success {
		fmt.Println(result.Stdout)
		return nil
	}

	return fmt.Errorf("Docker is not running")
}