package mysql

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string                                   { return "mysql" }
func (p *Plugin) Description() string                            { return "Manage MySQL/MariaDB Database Server" }
func (p *Plugin) Author() string                                 { return "VPS-Init" }
func (p *Plugin) Version() string                                { return "0.0.1" }
func (p *Plugin) Initialize(config map[string]interface{}) error { return nil }

// Enhanced plugin interface methods
func (p *Plugin) Validate() error {
	// MySQL plugin validation logic
	return nil
}

func (p *Plugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{}
}

func (p *Plugin) Compatibility() plugin.Compatibility {
	return plugin.Compatibility{
		MinVPSInitVersion: "1.0.0",
		GoVersion:         "1.19",
		Platforms:         []string{"linux/amd64", "linux/arm64"},
		Tags:              []string{"database", "mysql", "mariadb"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "mysql",
		Description: "Manage MySQL/MariaDB Database Server",
		Version:     "0.0.1",
		Author:      "VPS-Init",
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-init-plugins/mysql",
		Tags:        []string{"database", "mysql", "mariadb"},
		Validated:   true,
		TrustLevel:  "official",
		BuildInfo: plugin.BuildInfo{
			GoVersion: "1.21",
		},
	}
}

func (p *Plugin) Start(ctx context.Context) error                { return nil }
func (p *Plugin) Stop(ctx context.Context) error                 { return nil }
func (p *Plugin) GetRootCommand() *cobra.Command                 { return nil }

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install MariaDB Server and secure it",
			Handler:     p.installHandler,
		},
		{
			Name:        "create-db",
			Description: "Create a new database",
			Handler:     p.createDbHandler,
		},
		{
			Name:        "create-user",
			Description: "Create a new database user",
			Handler:     p.createUserHandler,
		},
		{
			Name:        "grant",
			Description: "Grant privileges to a user on a database",
			Handler:     p.grantHandler,
		},
		{
			Name:        "status",
			Description: "Check service status",
			Handler:     p.statusHandler,
		},
	}
}

// Handlers

func (p *Plugin) installHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🗄️  Installing MariaDB Server...")
	pass := getSudoPass(flags)

	// Update
	if res := conn.RunSudo("apt-get update", pass); !res.Success {
		return fmt.Errorf("apt update failed: %s", res.Stderr)
	}

	// Install
	if res := conn.RunSudo("apt-get install -y mariadb-server", pass); !res.Success {
		return fmt.Errorf("installation failed: %s", res.Stderr)
	}

	// Secure Installation
	// We'll do a basic automated security setup using SQL commands since 'mysql_secure_installation' is interactive.
	fmt.Println("🔒 Securing MariaDB...")

	// Commands to lock down root, remove anon users, remove test db
	secureSql := `
DELETE FROM mysql.user WHERE User='';
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test_%';
FLUSH PRIVILEGES;
`
	// Write to tmp file
	conn.WriteFile(secureSql, "/tmp/secure_mysql.sql")

	// Execute as root
	if res := conn.RunSudo("mysql -u root < /tmp/secure_mysql.sql", pass); !res.Success {
		// Verify if it failed because it's already secured (maybe root has password now?)
		// If it fails, log warning but continue
		fmt.Printf("Warning: automated security script had issues: %s\n", res.Stderr)
	}
	conn.RunSudo("rm /tmp/secure_mysql.sql", pass)

	fmt.Println("✅ MariaDB installed and secured.")
	return nil
}

func (p *Plugin) createDbHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: create-db <dbname>")
	}
	dbName := args[0]
	pass := getSudoPass(flags)

	fmt.Printf("Creating database %s...\n", dbName)
	cmd := fmt.Sprintf("mysql -u root -e 'CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;'", dbName)

	if res := conn.RunSudo(cmd, pass); !res.Success {
		return fmt.Errorf("failed to create db: %s", res.Stderr)
	}

	fmt.Printf("✅ Database %s created.\n", dbName)
	return nil
}

func (p *Plugin) createUserHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: create-user <username> <password>")
	}
	user := args[0]
	dbPass := args[1]
	pass := getSudoPass(flags)

	fmt.Printf("Creating user %s...\n", user)
	// Create user allowing connection from localhost
	cmd := fmt.Sprintf("mysql -u root -e \"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';\"", user, dbPass)

	if res := conn.RunSudo(cmd, pass); !res.Success {
		return fmt.Errorf("failed to create user: %s", res.Stderr)
	}

	fmt.Printf("✅ User %s created.\n", user)
	return nil
}

func (p *Plugin) grantHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: grant <username> <dbname>")
	}
	user := args[0]
	dbName := args[1]
	pass := getSudoPass(flags)

	fmt.Printf("Granting privileges to %s on %s...\n", user, dbName)
	cmd := fmt.Sprintf("mysql -u root -e \"GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost'; FLUSH PRIVILEGES;\"", dbName, user)

	if res := conn.RunSudo(cmd, pass); !res.Success {
		return fmt.Errorf("failed to grant privileges: %s", res.Stderr)
	}

	fmt.Println("✅ Privileges granted.")
	return nil
}

func (p *Plugin) statusHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	return conn.RunInteractive("systemctl status mariadb")
}

func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}
