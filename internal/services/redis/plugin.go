package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wasilwamark/vps-init/internal/distro"
	"github.com/wasilwamark/vps-init/internal/pkgmgr"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string {
	return "redis"
}

func (p *Plugin) Description() string {
	return "Redis database server management"
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Author() string {
	return "VPS-Init Team"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	return nil
}

func (p *Plugin) Validate() error {
	// Redis plugin validation logic
	return nil
}

func (p *Plugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{
		{
			Name:     "system",
			Version:  ">=1.0.0",
			Optional: false,
		},
	}
}

func (p *Plugin) Compatibility() plugin.Compatibility {
	return plugin.Compatibility{
		MinVPSInitVersion: "1.0.0",
		GoVersion:         "1.19",
		Platforms:         []string{"linux/amd64", "linux/arm64"},
		Tags:              []string{"database", "cache", "production-ready"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "redis",
		Description: "Redis database server management",
		Version:     "1.0.0",
		Author:      "VPS-Init Team",
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-redis",
		Tags:        []string{"database", "cache", "redis"},
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
			Description: "Install Redis server",
			Handler:     p.installHandler,
		},
		{
			Name:        "uninstall",
			Description: "Uninstall Redis server",
			Handler:     p.uninstallHandler,
		},
		{
			Name:        "start",
			Description: "Start Redis service",
			Handler:     p.startHandler,
		},
		{
			Name:        "stop",
			Description: "Stop Redis service",
			Handler:     p.stopHandler,
		},
		{
			Name:        "restart",
			Description: "Restart Redis service",
			Handler:     p.restartHandler,
		},
		{
			Name:        "status",
			Description: "Check Redis service status",
			Handler:     p.statusHandler,
		},
		{
			Name:        "configure",
			Description: "Configure Redis settings (interactive)",
			Handler:     p.configureHandler,
		},
		{
			Name:        "test",
			Description: "Test Redis connection",
			Handler:     p.testHandler,
		},
		{
			Name:        "info",
			Description: "Show Redis server information",
			Handler:     p.infoHandler,
		},
		{
			Name:        "backup",
			Description: "Backup Redis data",
			Handler:     p.backupHandler,
		},
	}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	return nil
}

func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

func (p *Plugin) installHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	sudoPass := getSudoPass(flags)

	fmt.Println("Installing Redis server...")

	// Check if Redis is already installed
	checkCmd := "redis-server --version"
	if result := conn.RunCommand(checkCmd, plugin.WithHideOutput()); result.Success {
		fmt.Println("Redis is already installed")
		return nil
	}

	// Update package list
	fmt.Println("Updating package list...")
	pkgMgr := getPackageManager(conn)
	updateCmd, _ := pkgMgr.Update()
	if result := conn.RunSudo(updateCmd, sudoPass); !result.Success {
		return fmt.Errorf("failed to update package list: %w", result.GetError())
	}

	// Install Redis
	fmt.Println("Installing Redis server...")
	installCmd, err := pkgMgr.Install("redis-server")
	if err != nil {
		return err
	}
	if result := conn.RunSudo(installCmd, sudoPass); !result.Success {
		return fmt.Errorf("failed to install Redis: %w", result.GetError())
	}

	// Enable Redis service
	fmt.Println("Enabling Redis service...")
	if result := conn.RunSudo("systemctl enable redis-server", sudoPass); !result.Success {
		return fmt.Errorf("failed to enable Redis service: %w", result.GetError())
	}

	fmt.Println("✅ Redis server installed successfully!")
	fmt.Println("You can now:")
	fmt.Println("  - Start Redis: vps-init redis start")
	fmt.Println("  - Configure Redis: vps-init redis configure")
	fmt.Println("  - Check status: vps-init redis status")

	return nil
}

func (p *Plugin) uninstallHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	sudoPass := getSudoPass(flags)

	fmt.Println("Uninstalling Redis server...")

	// Stop Redis service
	fmt.Println("Stopping Redis service...")
	conn.RunSudo("systemctl stop redis-server", sudoPass)

	// Disable Redis service
	fmt.Println("Disabling Redis service...")
	conn.RunSudo("systemctl disable redis-server", sudoPass)

	// Remove Redis package
	fmt.Println("Removing Redis server package...")
	pkgMgr := getPackageManager(conn)
	removeCmd, err := pkgMgr.Remove("redis-server", "redis-tools")
	if err != nil {
		return err
	}
	if result := conn.RunSudo(removeCmd, sudoPass); !result.Success {
		return fmt.Errorf("failed to remove Redis packages: %w", result.GetError())
	}

	// Remove Redis configuration and data directories
	fmt.Println("Removing Redis configuration and data...")
	conn.RunSudo("rm -rf /etc/redis", sudoPass)
	conn.RunSudo("rm -rf /var/lib/redis", sudoPass)
	conn.RunSudo("rm -rf /var/log/redis", sudoPass)

	fmt.Println("✅ Redis server uninstalled successfully!")

	return nil
}

func (p *Plugin) startHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	sudoPass := getSudoPass(flags)

	fmt.Println("Starting Redis service...")
	if result := conn.RunSudo("systemctl start redis-server", sudoPass); !result.Success {
		return fmt.Errorf("failed to start Redis service: %w", result.GetError())
	}

	fmt.Println("✅ Redis service started successfully!")
	return p.statusHandler(ctx, conn, args, flags)
}

func (p *Plugin) stopHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	sudoPass := getSudoPass(flags)

	fmt.Println("Stopping Redis service...")
	if result := conn.RunSudo("systemctl stop redis-server", sudoPass); !result.Success {
		return fmt.Errorf("failed to stop Redis service: %w", result.GetError())
	}

	fmt.Println("✅ Redis service stopped successfully!")
	return nil
}

func (p *Plugin) restartHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	sudoPass := getSudoPass(flags)

	fmt.Println("Restarting Redis service...")
	if result := conn.RunSudo("systemctl restart redis-server", sudoPass); !result.Success {
		return fmt.Errorf("failed to restart Redis service: %w", result.GetError())
	}

	fmt.Println("✅ Redis service restarted successfully!")
	return p.statusHandler(ctx, conn, args, flags)
}

func (p *Plugin) statusHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("Checking Redis service status...")

	// Check systemctl status
	if result := conn.RunCommand("systemctl is-active redis-server", plugin.WithHideOutput()); result.Success {
		fmt.Printf("🟢 Redis service status: %s\n", strings.TrimSpace(result.Stdout))
	} else {
		fmt.Printf("❌ Redis service is not active\n")
	}

	// Check if Redis is listening
	if result := conn.RunCommand("redis-cli ping 2>/dev/null || echo 'Not responding'", plugin.WithHideOutput()); result.Success {
		if strings.TrimSpace(result.Stdout) == "PONG" {
			fmt.Println("🟢 Redis server is responding")
		} else {
			fmt.Printf("🟡 Redis server status: %s\n", strings.TrimSpace(result.Stdout))
		}
	} else {
		fmt.Println("❌ Redis server is not responding")
	}

	// Show version if available
	if result := conn.RunCommand("redis-server --version 2>/dev/null || echo 'Version not available'", plugin.WithHideOutput()); result.Success {
		fmt.Printf("📦 Version: %s\n", strings.TrimSpace(result.Stdout))
	}

	return nil
}

func (p *Plugin) configureHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {

	configureScript := `
echo "🔧 Redis Configuration Menu"
echo "This will help you configure Redis server settings."
echo ""
echo "Current configuration file: /etc/redis/redis.conf"
echo ""
read -p "Do you want to view current configuration? (y/n): " view_config
if [[ $view_config =~ ^[Yy]$ ]]; then
    sudo grep -E "^(port|bind|requirepass|maxmemory|save)" /etc/redis/redis.conf || echo "Configuration not found"
fi
echo ""
read -p "Set Redis port (default: 6379): " redis_port
redis_port=${redis_port:-6379}
echo ""
read -p "Bind to localhost only? (Y/n): " bind_localhost
if [[ ! $bind_localhost =~ ^[Nn]$ ]]; then
    redis_bind="127.0.0.1"
else
    read -p "Enter bind address (e.g., 0.0.0.0): " redis_bind
    redis_bind=${redis_bind:-0.0.0.0}
fi
echo ""
read -p "Set password? (y/N): " set_password
if [[ $set_password =~ ^[Yy]$ ]]; then
    read -s -p "Enter Redis password: " redis_password
    echo
fi
echo ""
echo "Applying configuration..."
sudo cp /etc/redis/redis.conf /etc/redis/redis.conf.backup
sudo sed -i "s/^port .*/port $redis_port/" /etc/redis/redis.conf
sudo sed -i "s/^bind .*/bind $redis_bind/" /etc/redis/redis.conf
if [[ -n $redis_password ]]; then
    sudo sed -i "s/# requirepass .*/requirepass $redis_password/" /etc/redis/redis.conf
fi
echo "✅ Configuration updated!"
echo "Restarting Redis service..."
sudo systemctl restart redis-server
echo "✅ Redis service restarted!"
`
	return conn.RunInteractive(configureScript)
}

func (p *Plugin) testHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("Testing Redis connection...")

	// Test basic connectivity
	if result := conn.RunCommand("redis-cli ping 2>/dev/null", plugin.WithHideOutput()); result.Success {
		if strings.TrimSpace(result.Stdout) == "PONG" {
			fmt.Println("✅ Redis server is responding!")
		} else {
			fmt.Printf("🟡 Unexpected response: %s\n", result.Stdout)
		}
	} else {
		fmt.Println("❌ Failed to connect to Redis server")
		fmt.Println("Make sure Redis is running: vps-init redis start")
		return fmt.Errorf("Redis server not responding")
	}

	// Test basic operations
	fmt.Println("\nTesting basic Redis operations...")

	// Test SET operation
	if result := conn.RunCommand("redis-cli set vps_init_test 'Hello from VPS-Init' 2>/dev/null", plugin.WithHideOutput()); result.Success {
		fmt.Println("✅ SET operation successful")
	} else {
		fmt.Printf("❌ SET operation failed\n")
	}

	// Test GET operation
	if result := conn.RunCommand("redis-cli get vps_init_test 2>/dev/null", plugin.WithHideOutput()); result.Success {
		fmt.Printf("✅ GET operation successful: %s\n", strings.TrimSpace(result.Stdout))
	} else {
		fmt.Printf("❌ GET operation failed\n")
	}

	// Clean up test key
	conn.RunCommand("redis-cli del vps_init_test 2>/dev/null", plugin.WithHideOutput())

	// Show Redis info
	fmt.Println("\nRedis Server Information:")
	if result := conn.RunCommand("redis-cli info server | head -10", plugin.WithHideOutput()); result.Success {
		fmt.Print(result.Stdout)
	}

	return nil
}

func (p *Plugin) infoHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("Redis Server Information:")
	fmt.Println("========================")

	// Get Redis server info
	if result := conn.RunCommand("redis-cli info server 2>/dev/null || echo 'Redis server not accessible'", plugin.WithHideOutput()); result.Success {
		if result.Stdout != "" {
			// Parse and display key information
			lines := strings.Split(result.Stdout, "\n")
			for _, line := range lines {
				if strings.Contains(line, "redis_version:") {
					fmt.Printf("Version: %s\n", strings.TrimPrefix(line, "redis_version:"))
				} else if strings.Contains(line, "redis_mode:") {
					fmt.Printf("Mode: %s\n", strings.TrimPrefix(line, "redis_mode:"))
				} else if strings.Contains(line, "os:") {
					fmt.Printf("OS: %s\n", strings.TrimPrefix(line, "os:"))
				} else if strings.Contains(line, "arch_bits:") {
					fmt.Printf("Architecture: %s bits\n", strings.TrimPrefix(line, "arch_bits:"))
				} else if strings.Contains(line, "uptime_in_seconds:") {
					fmt.Printf("Uptime: %s seconds\n", strings.TrimPrefix(line, "uptime_in_seconds:"))
				}
			}
		}
	} else {
		fmt.Println("❌ Cannot connect to Redis server")
		return fmt.Errorf("Redis server not accessible")
	}

	fmt.Println("\nMemory Usage:")
	if result := conn.RunCommand("redis-cli info memory 2>/dev/null | grep -E '(used_memory_human|maxmemory_human)' || echo 'Memory info not available'", plugin.WithHideOutput()); result.Success {
		fmt.Print(result.Stdout)
	}

	fmt.Println("\nConnected Clients:")
	if result := conn.RunCommand("redis-cli info clients 2>/dev/null | grep connected_clients || echo 'Client info not available'", plugin.WithHideOutput()); result.Success {
		fmt.Print(result.Stdout)
	}

	fmt.Println("\nDatabase Statistics:")
	if result := conn.RunCommand("redis-cli info keyspace 2>/dev/null || echo 'Keyspace info not available'", plugin.WithHideOutput()); result.Success {
		if strings.TrimSpace(result.Stdout) == "" {
			fmt.Println("No database statistics available")
		} else {
			fmt.Print(result.Stdout)
		}
	}

	return nil
}

func (p *Plugin) backupHandler(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
	sudoPass := getSudoPass(flags)

	fmt.Println("Creating Redis backup...")

	// Create backup directory
	backupDir := "/var/backups/redis"
	fmt.Printf("Creating backup directory: %s\n", backupDir)
	if result := conn.RunSudo(fmt.Sprintf("mkdir -p %s", backupDir), sudoPass); !result.Success {
		return fmt.Errorf("failed to create backup directory: %w", result.GetError())
	}

	// Generate timestamp
	result := conn.RunCommand("date '+%Y%m%d_%H%M%S'", plugin.WithHideOutput())
	if !result.Success {
		return fmt.Errorf("failed to generate timestamp")
	}
	timestamp := strings.TrimSpace(result.Stdout)

	// Create backup
	backupFile := fmt.Sprintf("%s/redis_backup_%s.rdb", backupDir, timestamp)
	fmt.Printf("Creating backup file: %s\n", backupFile)

	// Use BGSAVE for non-blocking backup
	fmt.Println("Triggering background save...")
	if result := conn.RunCommand("redis-cli BGSAVE", plugin.WithHideOutput()); !result.Success {
		return fmt.Errorf("failed to trigger background save")
	}

	// Wait for backup to complete
	fmt.Println("Waiting for backup to complete...")
	for i := 0; i < 30; i++ {
		if result := conn.RunCommand("redis-cli LASTSAVE", plugin.WithHideOutput()); result.Success {
			fmt.Printf("Backup created successfully at timestamp: %s\n", strings.TrimSpace(result.Stdout))
			break
		}
		fmt.Print(".")
	}

	// Copy the RDB file to backup location
	rdbPath := "/var/lib/redis/dump.rdb"
	copyCmd := fmt.Sprintf("sudo cp %s %s", rdbPath, backupFile)
	if result := conn.RunSudo(copyCmd, sudoPass); !result.Success {
		return fmt.Errorf("failed to copy RDB file: %w", result.GetError())
	}

	// Set proper permissions
	chmodCmd := fmt.Sprintf("sudo chmod 640 %s", backupFile)
	if result := conn.RunSudo(chmodCmd, sudoPass); !result.Success {
		return fmt.Errorf("failed to set backup file permissions: %w", result.GetError())
	}

	fmt.Printf("\n✅ Redis backup created successfully!\n")
	fmt.Printf("Backup file: %s\n", backupFile)
	if result := conn.RunCommand(fmt.Sprintf("sudo ls -lh %s | awk '{print $5}'", backupFile), plugin.WithHideOutput()); result.Success {
		fmt.Printf("Size: %s\n", strings.TrimSpace(result.Stdout))
	}

	return nil
}

// Helper function to get sudo password from flags
func getSudoPass(flags map[string]interface{}) string {
	if pass, ok := flags["sudo_password"].(string); ok {
		return pass
	}
	return ""
}

func getPackageManager(conn plugin.Connection) pkgmgr.PackageManager {
	distroInfo := conn.GetDistroInfo().(*distro.DistroInfo)
	return pkgmgr.GetPackageManager(distroInfo)
}
