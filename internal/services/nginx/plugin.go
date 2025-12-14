package nginx

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string {
	return "nginx"
}

func (p *Plugin) Description() string {
	return "Manage Nginx web server (install, config, ssl)"
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

func (p *Plugin) Dependencies() []string {
	return []string{} // No other plugin dependencies
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	return nil // No top-level command, just subcommands
}

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install Nginx",
			Handler:     p.installHandler,
		},
		{
			Name:        "status",
			Description: "Check Nginx status",
			Handler:     p.statusHandler,
		},
		{
			Name:        "start",
			Description: "Start Nginx",
			Handler:     p.serviceActionHandler("start"),
		},
		{
			Name:        "stop",
			Description: "Stop Nginx",
			Handler:     p.serviceActionHandler("stop"),
		},
		{
			Name:        "restart",
			Description: "Restart Nginx",
			Handler:     p.serviceActionHandler("restart"),
		},
		{
			Name:        "reload",
			Description: "Reload Nginx configuration",
			Handler:     p.serviceActionHandler("reload"),
		},
		{
			Name:        "logs",
			Description: "Stream Nginx logs",
			Handler:     p.logsHandler,
		},
		{
			Name:        "list-sites",
			Description: "List all configured sites",
			Handler:     p.listSitesHandler,
		},
		{
			Name:        "add-site",
			Description: "Add a new site (reverse proxy)",
			Handler:     p.addSiteHandler,
		},
		{
			Name:        "remove-site",
			Description: "Remove a site configuration",
			Handler:     p.removeSiteHandler,
		},
		{
			Name:        "install-ssl",
			Description: "Install SSL certificate using Certbot",
			Handler:     p.installSSLHandler,
		},
	}
}

func (p *Plugin) installHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🌐 Installing Nginx...")

	// 1. Update apt
	if res := conn.RunSudo("apt-get update", getSudoPass(flags)); !res.Success {
		return fmt.Errorf("failed to update apt: %s", res.Stderr)
	}

	// 2. Install nginx
	if res := conn.RunSudo("apt-get install -y nginx", getSudoPass(flags)); !res.Success {
		return fmt.Errorf("failed to install nginx: %s", res.Stderr)
	}

	fmt.Println("✅ Nginx installed successfully!")
	return nil
}

func (p *Plugin) statusHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	return conn.RunInteractive("systemctl status nginx")
}

func (p *Plugin) serviceActionHandler(action string) plugin.CommandHandler {
	return func(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
		pass := getSudoPass(flags)

		// For reload, always test config first
		if action == "reload" {
			fmt.Println("🔍 Testing Nginx configuration...")
			if res := conn.RunSudo("nginx -t", pass); !res.Success {
				return fmt.Errorf("nginx config test failed:\n%s", res.Stderr)
			}
		}

		fmt.Printf("⚙️  Running: systemctl %s nginx...\n", action)
		if res := conn.RunSudo(fmt.Sprintf("systemctl %s nginx", action), pass); !res.Success {
			return fmt.Errorf("failed to %s nginx: %s", action, res.Stderr)
		}
		fmt.Printf("✅ Nginx %sed successfully\n", action)
		return nil
	}
}

func (p *Plugin) logsHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("📜 Streaming Nginx logs (Ctrl+C to stop)...")
	// Using sudo for logs as accessing /var/log often requires root group or root
	// But journalctl might work if user is in systemd-journal group.
	// We'll try sudo interactive.

	cmd := "journalctl -u nginx -f"
	if pass := getSudoPass(flags); pass != "" {
		// If we have password, we can construct a sudo command.
		// BUT standard ssh RunInteractive relies on creating a PTY.
		// Passing password securely to sudo in interactive PTY is tricky without 'expect'.
		// So we will just run sudo and let the user type password if needed, OR
		// if we really want to auto-inject, we use the non-interactive RunSudo but that doesn't stream nicely?
		// Limitation: For interactive logs, we just let sudo ask for password if needed,
		// OR we trust the user has passwordless sudo for specific commands.
		// BETTER: Just run it. If it needs password, the PTY will show the prompt.
		cmd = "sudo journalctl -u nginx -f"
	} else {
		// Just run sudo, it might prompt
		cmd = "sudo journalctl -u nginx -f"
	}

	return conn.RunInteractive(cmd)
}

func (p *Plugin) listSitesHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔍 Fetching configured sites...")

	// List sites in sites-enabled
	res := conn.RunCommand("ls -1 /etc/nginx/sites-enabled/", false)
	if !res.Success {
		return fmt.Errorf("failed to list sites: %s", res.Stderr)
	}

	sites := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	if len(sites) == 0 || (len(sites) == 1 && sites[0] == "") {
		fmt.Println("No sites configured.")
		return nil
	}

	fmt.Println("\n📋 Configured Sites:")
	for _, site := range sites {
		if site == "" {
			continue
		}

		// Check if it's a symlink (enabled) or regular file
		checkRes := conn.RunCommand(fmt.Sprintf("test -L /etc/nginx/sites-enabled/%s && echo 'symlink' || echo 'file'", site), false)
		linkType := strings.TrimSpace(checkRes.Stdout)

		// Check if SSL is configured by looking for listen 443 in the config
		sslRes := conn.RunCommand(fmt.Sprintf("grep -q 'listen.*443' /etc/nginx/sites-enabled/%s && echo 'yes' || echo 'no'", site), false)
		hasSSL := strings.TrimSpace(sslRes.Stdout) == "yes"

		status := "✅"
		if linkType != "symlink" {
			status = "⚠️"
		}

		sslStatus := ""
		if hasSSL {
			sslStatus = " 🔒 SSL"
		}

		fmt.Printf("  %s %s%s\n", status, site, sslStatus)
	}

	return nil
}

func (p *Plugin) addSiteHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: add-site <domain> [--proxy <port>] [--file <local-path>] [--ssl]")
	}
	domain := args[0]
	// Basic parsing
	proxyPort := "3000" // default
	localConfigPath := ""
	ssl := false

	for i, arg := range args {
		if arg == "--proxy" && i+1 < len(args) {
			proxyPort = args[i+1]
		}
		if arg == "--file" && i+1 < len(args) {
			localConfigPath = args[i+1]
		}
		if arg == "--ssl" {
			ssl = true
		}
	}

	configContent := ""
	if localConfigPath != "" {
		fmt.Printf("📂 Reading local configuration from %s...\n", localConfigPath)
		content, err := os.ReadFile(localConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read local config file: %v", err)
		}
		configContent = string(content)
	} else {
		fmt.Printf("📝 Configuring site %s (proxying to localhost:%s)...\n", domain, proxyPort)
		configContent = fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://localhost:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
`, domain, proxyPort)
	}

	// Check if Nginx is installed
	if !conn.DirectoryExists("/etc/nginx/sites-available") {
		return fmt.Errorf("Nginx configuration directory not found. Is Nginx installed? Try running: vps-init <target> nginx install")
	}

	// Create temp file securely? Or just echo to path.
	// Since we need sudo to write to /etc/nginx, we write to /tmp first then move.
	tmpPath := fmt.Sprintf("/tmp/nginx_%s.conf", domain)
	if !conn.WriteFile(configContent, tmpPath) {
		return fmt.Errorf("failed to write temp config")
	}

	confPath := fmt.Sprintf("/etc/nginx/sites-available/%s", domain)

	// Move and Enable
	cmds := []string{
		fmt.Sprintf("mv %s %s", tmpPath, confPath),
		fmt.Sprintf("ln -sf %s /etc/nginx/sites-enabled/", confPath),
	}

	pass := getSudoPass(flags)
	for _, cmd := range cmds {
		if res := conn.RunSudo(cmd, pass); !res.Success {
			return fmt.Errorf("failed step '%s': %s", cmd, res.Stderr)
		}
	}

	// Verify Config with Rollback
	fmt.Println("🔍 Testing Nginx configuration...")
	if res := conn.RunSudo("nginx -t", pass); !res.Success {
		fmt.Printf("❌ Config test failed details:\n%s\n", res.Stderr)
		fmt.Println("🔄 Rolling back changes...")
		// Remove the symlink
		conn.RunSudo(fmt.Sprintf("rm -f /etc/nginx/sites-enabled/%s", domain), pass)
		return fmt.Errorf("nginx config test failed. Changes rolled back")
	}

	// Reload
	if res := conn.RunSudo("systemctl reload nginx", pass); !res.Success {
		return fmt.Errorf("failed to reload nginx: %s", res.Stderr)
	}

	fmt.Printf("✅ Site %s added and enabled!\n", domain)

	if ssl {
		fmt.Println("🔒 Proceeding to SSL installation...")
		return p.installSSLHandler(ctx, conn, []string{domain}, flags)
	}

	return nil
}

func (p *Plugin) removeSiteHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: remove-site <domain>")
	}
	domain := args[0]
	fmt.Printf("🗑️  Removing site %s...\n", domain)

	cmds := []string{
		fmt.Sprintf("rm -f /etc/nginx/sites-enabled/%s", domain),
		fmt.Sprintf("rm -f /etc/nginx/sites-available/%s", domain),
		"nginx -t", // Test config to make sure we didn't break anything (though removing shouldn't)
		"systemctl reload nginx",
	}

	pass := getSudoPass(flags)
	for _, cmd := range cmds {
		if res := conn.RunSudo(cmd, pass); !res.Success {
			return fmt.Errorf("failed step '%s': %s", cmd, res.Stderr)
		}
	}

	fmt.Printf("✅ Site %s removed successfully!\n", domain)
	return nil
}

func (p *Plugin) installSSLHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	domain := ""
	if len(args) > 0 {
		domain = args[0]
	} else {
		// Interactive selection
		fmt.Println("🔍 Fetching available sites...")
		res := conn.RunCommand("ls -1 /etc/nginx/sites-enabled/", false)
		if !res.Success {
			return fmt.Errorf("failed to list sites: %s", res.Stderr)
		}

		sites := strings.Split(strings.TrimSpace(res.Stdout), "\n")
		var validSites []string
		for _, s := range sites {
			if s != "" && s != "default" {
				validSites = append(validSites, s)
			}
		}

		if len(validSites) == 0 {
			return fmt.Errorf("no sites found in sites-enabled")
		}

		fmt.Println("\nAvailable sites:")
		for i, site := range validSites {
			fmt.Printf("  [%d] %s\n", i+1, site)
		}
		fmt.Println("  [0] Cancel")

		fmt.Print("\nSelect site to secure (enter number): ")
		var selection int
		_, err := fmt.Scanln(&selection)
		if err != nil || selection < 1 || selection > len(validSites) {
			if selection == 0 {
				return nil
			}
			return fmt.Errorf("invalid selection")
		}
		domain = validSites[selection-1]
	}

	pass := getSudoPass(flags)

	fmt.Println("🔒 Installing Certbot and SSL...")

	// Install certbot if needed
	installCmds := []string{
		"apt-get update",
		"apt-get install -y certbot python3-certbot-nginx",
	}
	for _, cmd := range installCmds {
		if res := conn.RunSudo(cmd, pass); !res.Success {
			// Don't error immediately, might be installed
		}
	}

	fmt.Printf("🔐 Obtaining certificate for %s...\n", domain)
	// Run certbot
	cmd := fmt.Sprintf("sudo certbot --nginx -d %s", domain)
	return conn.RunInteractive(cmd)
}

// Helper
func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}
