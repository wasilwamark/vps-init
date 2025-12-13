package alias

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/config"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct {
	config map[string]interface{}
}

func NewPlugin() plugin.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "alias"
}

func (p *Plugin) Description() string {
	return "Server alias management"
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
			Name:        "add",
			Description: "Add a server alias",
			Args: []plugin.Argument{
				{
					Name:        "name",
					Description: "Alias name",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
				{
					Name:        "connection",
					Description: "User@host connection string",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleAdd,
		},
		{
			Name:        "list",
			Description: "List all server aliases",
			Handler:     p.handleList,
		},
		{
			Name:        "remove",
			Description: "Remove a server alias",
			Args: []plugin.Argument{
				{
					Name:        "name",
					Description: "Alias name to remove",
					Required:    true,
					Type:        plugin.ArgumentTypeString,
				},
			},
			Handler: p.handleRemove,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage server aliases",
		Long: `Manage server aliases for easier access.

Examples:
  vps-init alias add myserver user@host.com
  vps-init alias list
  vps-init alias remove myserver`,
	}

	addCmd := &cobra.Command{
		Use:   "add <name> <user@host>",
		Short: "Add a server alias",
		Args:  cobra.ExactArgs(2),
		Run:   p.runAdd,
	}
	cmd.AddCommand(addCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all server aliases",
		Run:   p.runList,
	}
	cmd.AddCommand(listCmd)

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a server alias",
		Args:  cobra.ExactArgs(1),
		Run:   p.runRemove,
	}
	cmd.AddCommand(removeCmd)

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

// Command handlers
func (p *Plugin) handleAdd(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("name and connection are required")
	}

	cfg := config.New()
	if err := cfg.SetAlias(args[0], args[1]); err != nil {
		return fmt.Errorf("failed to add alias: %w", err)
	}

	fmt.Printf("✅ Added alias '%s' for %s\n", args[0], args[1])
	return nil
}

func (p *Plugin) handleList(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	cfg := config.New()
	aliases := cfg.GetAliases()

	if len(aliases) == 0 {
		fmt.Println("No aliases found. Use 'vps-init alias add' to add one.")
		return nil
	}

	fmt.Println("Server Aliases:")
	for name, connection := range aliases {
		fmt.Printf("  %s: %s\n", name, connection)
	}
	return nil
}

func (p *Plugin) handleRemove(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("alias name is required")
	}

	cfg := config.New()
	if err := cfg.RemoveAlias(args[0]); err != nil {
		return fmt.Errorf("failed to remove alias: %w", err)
	}

	fmt.Printf("✅ Removed alias '%s'\n", args[0])
	return nil
}

// Cobra command runners
func (p *Plugin) runAdd(cmd *cobra.Command, args []string) {
	cfg := config.New()
	if err := cfg.SetAlias(args[0], args[1]); err != nil {
		fmt.Printf("❌ Failed to add alias: %v\n", err)
		return
	}
	fmt.Printf("✅ Added alias '%s' for %s\n", args[0], args[1])
}

func (p *Plugin) runList(cmd *cobra.Command, args []string) {
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
}

func (p *Plugin) runRemove(cmd *cobra.Command, args []string) {
	cfg := config.New()
	if err := cfg.RemoveAlias(args[0]); err != nil {
		fmt.Printf("❌ Failed to remove alias: %v\n", err)
		return
	}
	fmt.Printf("✅ Removed alias '%s'\n", args[0])
}