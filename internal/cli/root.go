package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/config"
	"github.com/wasilwamark/vps-init/internal/services/nginx"
	"github.com/wasilwamark/vps-init/internal/services/docker"
	"github.com/wasilwamark/vps-init/internal/services/monitoring"
)

var rootCmd = &cobra.Command{
	Use:   "vps-init",
	Short: "VPS-Init - Configure your servers with simple commands",
	Long: `VPS-Init is a CLI tool that makes server configuration easy.

Examples:
  vps-init mark@1.2.3.4 nginx install
  vps-init mark@1.2.3.4 nginx install-ssl api.tiza.africa
  vps-init myserver docker install
  vps-init --add-alias myserver mark@1.2.3.4

Use "vps-init help" for more information.`,
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage server aliases",
}

var addAliasCmd = &cobra.Command{
	Use:   "add <name> <user@host>",
	Short: "Add a server alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		if err := cfg.SetAlias(args[0], args[1]); err != nil {
			fmt.Printf("❌ Failed to add alias: %v\n", err)
			return
		}
		fmt.Printf("✅ Added alias '%s' for %s\n", args[0], args[1])
	},
}

var listAliasesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all server aliases",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		aliases := cfg.GetAliases()

		if len(aliases) == 0 {
			fmt.Println("No aliases found. Use 'vps-init alias add' to add one.")
			return
		}

		fmt.Println("Server Aliases:")
		for name, connection := range aliases {
			fmt.Printf("  %s: %s\n", name, connection)
		}
	},
}

var removeAliasCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a server alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		if err := cfg.RemoveAlias(args[0]); err != nil {
			fmt.Printf("❌ Failed to remove alias: %v\n", err)
			return
		}
		fmt.Printf("✅ Removed alias '%s'\n", args[0])
	},
}

func init() {
	// Add alias commands
	aliasCmd.AddCommand(addAliasCmd)
	aliasCmd.AddCommand(listAliasesCmd)
	aliasCmd.AddCommand(removeAliasCmd)
	rootCmd.AddCommand(aliasCmd)

	// Add service commands
	rootCmd.AddCommand(nginx.Command())
	rootCmd.AddCommand(docker.Command())
	rootCmd.AddCommand(monitoring.Command())

	// Support direct command execution (vps-init user@host service command)
	if len(os.Args) > 1 && os.Args[1] != "alias" && os.Args[1] != "help" && os.Args[1] != "--help" && os.Args[1] != "nginx" && os.Args[1] != "docker" && os.Args[1] != "monitoring" {
		// Direct execution mode
		executeDirectCommand()
	}
}

func executeDirectCommand() {
	if len(os.Args) < 4 {
		fmt.Println(rootCmd.Long)
		os.Exit(1)
	}

	target := os.Args[1]
	service := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:]

	// Route to appropriate service
	switch service {
	case "nginx":
		nginx.ExecuteDirect(target, command, args)
	case "docker":
		docker.ExecuteDirect(target, command, args)
	case "monitoring":
		monitoring.ExecuteDirect(target, command, args)
	default:
		fmt.Printf("❌ Unknown service: %s\n", service)
		fmt.Println("Available services: nginx, docker, monitoring")
		os.Exit(1)
	}

	os.Exit(0)
}

