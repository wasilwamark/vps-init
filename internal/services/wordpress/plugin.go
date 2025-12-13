package wordpress

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string                                   { return "wordpress" }
func (p *Plugin) Description() string                            { return "WordPress Manager (LEMP Stack)" }
func (p *Plugin) Author() string                                 { return "VPS-Init" }
func (p *Plugin) Version() string                                { return "0.0.1" }
func (p *Plugin) Dependencies() []string                         { return []string{"mysql", "nginx"} }
func (p *Plugin) Initialize(config map[string]interface{}) error { return nil }
func (p *Plugin) Start(ctx context.Context) error                { return nil }
func (p *Plugin) Stop(ctx context.Context) error                 { return nil }
func (p *Plugin) GetRootCommand() *cobra.Command                 { return nil }

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install PHP and WP-CLI dependencies",
			Handler:     p.installHandler,
		},
		{
			Name:        "create-site",
			Description: "Deploy a new WordPress site (Interactive Wizard)",
			Handler:     p.createSiteHandler,
		},
	}
}

// Handlers

func (p *Plugin) installHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🐘 Installing PHP and Dependencies...")
	pass := getSudoPass(flags)

	// Update
	conn.RunSudo("apt-get update", pass)

	// Install PHP (and common Extensions), Curl, Unzip
	pkgs := "php-fpm php-mysql php-curl php-gd php-mbstring php-xml php-xmlrpc php-soap php-intl php-zip unzip curl"
	if res := conn.RunSudo(fmt.Sprintf("apt-get install -y %s", pkgs), pass); !res.Success {
		return fmt.Errorf("php install failed: %s", res.Stderr)
	}

	fmt.Println("🛠️  Installing WP-CLI...")
	// Download WP-CLI
	conn.RunSudo("curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar", pass)
	conn.RunSudo("chmod +x wp-cli.phar", pass)
	conn.RunSudo("mv wp-cli.phar /usr/local/bin/wp", pass)

	// Verify
	if res := conn.RunCommand("wp --info", false); !res.Success {
		fmt.Printf("Warning: WP-CLI install verification failed: %s\n", res.Stderr)
	} else {
		fmt.Println("✅ WP-CLI installed.")
	}

	fmt.Println("✅ WordPress Environment Ready.")
	return nil
}

func (p *Plugin) createSiteHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: create-site <domain>")
	}
	domain := args[0]
	pass := getSudoPass(flags)

	// Interactive Wizard
	fmt.Println("🚀 Standard WordPress Deployment Wizard")
	fmt.Printf("Domain: %s\n", domain)

	// Input gathering
	// Use simplified flow for Agent: predefined or auto-gen values if not interactive?
	// The SSH RunInteractive doesn't support easy bi-directional variable capture from user input on remote.
	// We must ask user on LOCAL side.

	// Assuming local interactive inputs
	var dbName, dbUser, dbPass, adminUser, adminPass, adminEmail string

	fmt.Printf("Database Name [wp_%s]: ", strings.ReplaceAll(domain, ".", "_"))
	fmt.Scanln(&dbName)
	if dbName == "" {
		dbName = fmt.Sprintf("wp_%s", strings.ReplaceAll(domain, ".", "_"))
	}

	fmt.Printf("Database User [user_%s]: ", dbName)
	fmt.Scanln(&dbUser)
	if dbUser == "" {
		dbUser = fmt.Sprintf("user_%s", dbName)
	}

	fmt.Print("Database Password: ")
	fmt.Scanln(&dbPass)
	if dbPass == "" {
		return fmt.Errorf("password required")
	}

	fmt.Print("WP Admin User [admin]: ")
	fmt.Scanln(&adminUser)
	if adminUser == "" {
		adminUser = "admin"
	}

	fmt.Print("WP Admin Password: ")
	fmt.Scanln(&adminPass)
	if adminPass == "" {
		return fmt.Errorf("password required")
	}

	fmt.Print("WP Admin Email: ")
	fmt.Scanln(&adminEmail)
	if adminEmail == "" {
		return fmt.Errorf("email required")
	}

	webRoot := fmt.Sprintf("/var/www/%s", domain)

	// 1. Create Database & User
	fmt.Println("\n🗄️  Configuring Database...")
	cmds := []string{
		fmt.Sprintf("mysql -u root -e 'CREATE DATABASE IF NOT EXISTS %s;'", dbName),
		fmt.Sprintf("mysql -u root -e \"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';\"", dbUser, dbPass),
		fmt.Sprintf("mysql -u root -e \"GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost'; FLUSH PRIVILEGES;\"", dbName, dbUser),
	}
	for _, cmd := range cmds {
		if res := conn.RunSudo(cmd, pass); !res.Success {
			return fmt.Errorf("db step failed: %s", res.Stderr)
		}
	}

	// 2. Setup Web Root
	fmt.Println("📂 Setting up Web Root...")
	conn.RunSudo(fmt.Sprintf("mkdir -p %s", webRoot), pass)
	// Temporarily own by current user or root for WP-CLI operations, later www-data
	// Running WP-CLI as root requires --allow-root

	// 3. Download WordPress
	fmt.Println("⬇️  Downloading WordPress...")
	if res := conn.RunSudo(fmt.Sprintf("wp core download --path=%s --allow-root", webRoot), pass); !res.Success {
		return fmt.Errorf("wp download failed: %s", res.Stderr)
	}

	// 4. Create Config
	fmt.Println("⚙️  Configuring wp-config.php...")
	confCmd := fmt.Sprintf("wp config create --dbname=%s --dbuser=%s --dbpass='%s' --path=%s --allow-root", dbName, dbUser, dbPass, webRoot)
	if res := conn.RunSudo(confCmd, pass); !res.Success {
		return fmt.Errorf("wp config failed: %s", res.Stderr)
	}

	// 5. Install WordPress
	fmt.Println("💿 Installing WordPress Core...")
	instCmd := fmt.Sprintf("wp core install --url=http://%s --title='%s' --admin_user=%s --admin_password='%s' --admin_email=%s --path=%s --allow-root",
		domain, domain, adminUser, adminPass, adminEmail, webRoot)
	if res := conn.RunSudo(instCmd, pass); !res.Success {
		return fmt.Errorf("wp install failed: %s", res.Stderr)
	}

	// 6. Permissions
	fmt.Println("🔒 Setting Permissions...")
	conn.RunSudo(fmt.Sprintf("chown -R www-data:www-data %s", webRoot), pass)
	conn.RunSudo(fmt.Sprintf("chmod -R 755 %s", webRoot), pass)

	// 7. Nginx Config
	fmt.Println("🌐 Configuring Nginx...")
	// We need to determine PHP socket path. Usually /run/php/php8.1-fpm.sock or similar.
	// Let's try to find it.
	sockRes := conn.RunCommand("find /run/php -name 'php*-fpm.sock' | head -n 1", false)
	phpSock := strings.TrimSpace(sockRes.Stdout)
	if phpSock == "" {
		phpSock = "unix:/var/run/php/php-fpm.sock" // fallback
	} else {
		phpSock = "unix:" + phpSock
	}

	nginxConf := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    root %s;
    index index.php index.html index.htm;

    location / {
        try_files $uri $uri/ /index.php?$args;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass %s;
    }

    location ~ /\.ht {
        deny all;
    }
}
`, domain, webRoot, phpSock)

	tmpNginx := fmt.Sprintf("/tmp/nginx_%s", domain)
	conn.WriteFile(nginxConf, tmpNginx)

	conn.RunSudo(fmt.Sprintf("mv %s /etc/nginx/sites-available/%s", tmpNginx, domain), pass)
	conn.RunSudo(fmt.Sprintf("ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/", domain, domain), pass)

	// Test & Reload Nginx
	if res := conn.RunSudo("nginx -t", pass); !res.Success {
		// Rollback symlink
		conn.RunSudo(fmt.Sprintf("rm /etc/nginx/sites-enabled/%s", domain), pass)
		return fmt.Errorf("nginx config failed: %s", res.Stderr)
	}
	conn.RunSudo("systemctl reload nginx", pass)

	fmt.Printf("\n✅ WordPress Site http://%s deployed successfully!\n", domain)
	return nil
}

func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}
