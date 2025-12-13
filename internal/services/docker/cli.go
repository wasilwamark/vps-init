package docker

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/config"
	"github.com/wasilwamark/vps-init/internal/ssh"
)

// Command returns the cobra command for docker
func Command() *cobra.Command {
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
		Run:   runInstall,
	}
	cmd.AddCommand(installCmd)

	return cmd
}

// ExecuteDirect handles direct command execution
func ExecuteDirect(target, command string, args []string) {
	cfg := config.New()
	connectionString := cfg.ResolveTarget(target)

	parts := strings.Split(connectionString, "@")
	if len(parts) != 2 {
		fmt.Printf("❌ Invalid target format: %s\n", target)
		os.Exit(1)
	}

	conn := ssh.New(parts[1], parts[0])
	service := New(conn)

	switch command {
	case "install":
		if !service.Install() {
			os.Exit(1)
		}
	case "status":
		if err := service.Status(); err != nil {
			os.Exit(1)
		}
	case "deploy":
		composeFile := "docker-compose.yml"
		if len(args) > 0 {
			composeFile = args[0]
		}
		if !service.Deploy(composeFile) {
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ Unknown docker command: %s\n", command)
		os.Exit(1)
	}
}

func runInstall(cmd *cobra.Command, args []string) {
	fmt.Println("Use: vps-init <user@host> docker install")
}