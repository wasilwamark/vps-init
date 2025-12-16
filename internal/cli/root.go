package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/config"
	vps_ssh "github.com/wasilwamark/vps-init-ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

var rootCmd = &cobra.Command{
	Use:     "vps-init",
	Version: "0.0.1",
	Short:   "VPS-Init - Configure your servers with simple commands",
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
	Example: `  vps-init alias add ovh ubuntu@1.2.3.4
  vps-init alias add ovh ubuntu@1.2.3.4 --sudo-password 'my-secret'`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		if err := cfg.SetAlias(args[0], args[1]); err != nil {
			fmt.Printf("❌ Failed to add alias: %v\n", err)
			return
		}

		// Handle sudo password if flag set
		sudoPass, _ := cmd.Flags().GetString("sudo-password")
		if sudoPass != "" {
			if err := cfg.SetSecret(args[0], sudoPass); err != nil {
				fmt.Printf("⚠️  Alias added, but failed to save sudo password: %v\n", err)
			} else {
				fmt.Printf("✅ Added alias '%s' for %s (with sudo password saved)\n", args[0], args[1])
				return
			}
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
	addAliasCmd.Flags().String("sudo-password", "", "Optional sudo password for the server")
	aliasCmd.AddCommand(addAliasCmd)
	aliasCmd.AddCommand(listAliasesCmd)
	aliasCmd.AddCommand(removeAliasCmd)
	rootCmd.AddCommand(aliasCmd)

}

func executeDirectCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vps-init user@host <plugin> <command> [args...]")
		os.Exit(1)
	}

	target := os.Args[1]

	// Resolve alias if present
	cfg := config.New()
	target = cfg.ResolveTarget(target)

	pluginName := os.Args[2]

	// Default to "help" or equivalent if no command provided?
	// The current signature expects at least 4 args provided in valid check
	// But let's be more flexible.
	cmdName := ""
	var args []string
	if len(os.Args) > 3 {
		cmdName = os.Args[3]
		args = os.Args[4:]
	}

	// Get registry
	registry := plugin.GetBuiltinRegistry()

	// Find plugin
	pl, exists := registry.Get(pluginName)
	if !exists {
		// Try to see if it's an alias command or something else?
		// Actually, aliases are handled by config.
		// For now just error
		fmt.Printf("❌ Unknown service/plugin: %s\n", pluginName)
		fmt.Println("Run 'vps-init plugin list' to see available plugins.")
		os.Exit(1)
	}

	// Find command
	var commandToRun *plugin.Command
	for _, cmd := range pl.GetCommands() {
		if cmd.Name == cmdName {
			commandToRun = &cmd
			break
		}
	}

	if commandToRun == nil {
		fmt.Printf("❌ Unknown command '%s' for plugin '%s'\n", cmdName, pluginName)
		fmt.Println("Available commands:")
		for _, cmd := range pl.GetCommands() {
			fmt.Printf("  %s: %s\n", cmd.Name, cmd.Description)
		}
		os.Exit(1)
	}

	// Establish SSH connection
	// We need to parse target (user@host)
	parts := strings.Split(target, "@")
	if len(parts) != 2 {
		fmt.Printf("❌ Invalid target format '%s'. Expected 'user@host' or a valid alias.\n", target)
		fmt.Println("Tip: Use 'vps-init alias list' to see available aliases.")
		os.Exit(1)
	}
	user := parts[0]
	host := parts[1]

	ctx := context.Background()
	conn := vps_ssh.New(host, user)
	// vps_ssh.New does not return error, it just creates the struct
	defer conn.Disconnect()

	if !conn.Connect() {
		fmt.Printf("❌ Failed to establish SSH connection manually (Connect returned false)\n")
		os.Exit(1)
	}

	// Execute handler
	// Parse args for flags
	flags := make(map[string]interface{})

	// Read sudo password from environment for security (Per Alias)
	// We check if the original target was an alias
	if _, isAlias := cfg.GetAlias(os.Args[1]); isAlias {
		aliasName := strings.ToUpper(os.Args[1])
		// Normalize alias name for env var (e.g. replace - with _)
		aliasName = strings.ReplaceAll(aliasName, "-", "_")
		envVar := fmt.Sprintf("SSH_SUDO_PWD_%s", aliasName)

		if envPass := os.Getenv(envVar); envPass != "" {
			flags["sudo-password"] = envPass
		} else {
			// Check local secrets store
			if secret, exists := cfg.GetSecret(os.Args[1]); exists {
				flags["sudo-password"] = secret
			}
		}
	}

	if err := commandToRun.Handler(ctx, conn, args, flags); err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
		os.Exit(1)
	}
}
