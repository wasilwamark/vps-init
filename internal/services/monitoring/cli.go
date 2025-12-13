package monitoring

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/config"
	"github.com/wasilwamark/vps-init/internal/ssh"
)

// Command returns the cobra command for monitoring
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitoring",
		Short: "System monitoring setup",
		Long: `Setup system monitoring and alerting.

Examples:
  vps-init monitoring install`,
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install monitoring tools",
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
	default:
		fmt.Printf("❌ Unknown monitoring command: %s\n", command)
		os.Exit(1)
	}
}

func runInstall(cmd *cobra.Command, args []string) {
	fmt.Println("Use: vps-init <user@host> monitoring install")
}