package nginx

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/config"
	"github.com/wasilwamark/vps-init/internal/ssh"
)

// Command returns the cobra command for nginx
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nginx",
		Short: "Nginx web server management",
		Long: `Manage Nginx web server configuration.

Examples:
  vps-init nginx install
  vps-init nginx install-ssl domain.com`,
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install Nginx web server",
		Run:   runInstall,
	}
	cmd.AddCommand(installCmd)

	installSSLCmd := &cobra.Command{
		Use:   "install-ssl [domain]",
		Short: "Install Nginx with SSL certificate",
		Args:  cobra.ExactArgs(1),
		Run:   runInstallSSL,
	}
	cmd.AddCommand(installSSLCmd)

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
	case "install-ssl":
		if len(args) == 0 {
			fmt.Println("❌ Domain is required for SSL installation")
			os.Exit(1)
		}
		if !service.InstallWithSSL(args[0]) {
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ Unknown nginx command: %s\n", command)
		os.Exit(1)
	}
}

func runInstall(cmd *cobra.Command, args []string) {
	fmt.Println("Use: vps-init <user@host> nginx install")
}

func runInstallSSL(cmd *cobra.Command, args []string) {
	fmt.Println("Use: vps-init <user@host> nginx install-ssl <domain>")
}