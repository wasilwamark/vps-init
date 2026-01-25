package keycloak

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
	return "keycloak"
}

func (p *Plugin) Description() string {
	return "Keycloak identity and access management service"
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
	return nil
}

func (p *Plugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{
		{
			Name:     "docker",
			Version:  ">=1.0.0",
			Optional: false,
		},
		{
			Name:     "nginx",
			Version:  ">=1.0.0",
			Optional: false,
		},
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
		Tags:              []string{"identity", "authentication", "sso", "enterprise"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "keycloak",
		Description: "Keycloak identity and access management service",
		Version:     "1.0.0",
		Author:      "VPS-Init Team",
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-init/services/keycloak",
		Tags:        []string{"identity", "authentication", "sso", "keycloak"},
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
			Description: "Install Keycloak with Docker and PostgreSQL",
			Handler:     p.installHandler,
		},
		{
			Name:        "uninstall",
			Description: "Remove Keycloak installation",
			Handler:     p.uninstallHandler,
		},
		{
			Name:        "start",
			Description: "Start Keycloak services",
			Handler:     p.serviceActionHandler("start"),
		},
		{
			Name:        "stop",
			Description: "Stop Keycloak services",
			Handler:     p.serviceActionHandler("stop"),
		},
		{
			Name:        "restart",
			Description: "Restart Keycloak services",
			Handler:     p.serviceActionHandler("restart"),
		},
		{
			Name:        "status",
			Description: "Check Keycloak service status",
			Handler:     p.statusHandler,
		},
		{
			Name:        "logs",
			Description: "View Keycloak service logs",
			Handler:     p.logsHandler,
		},
		{
			Name:        "realm",
			Description: "Manage Keycloak realms (create/list/delete)",
			Handler:     p.realmHandler,
		},
		{
			Name:        "user",
			Description: "Manage Keycloak users (create/list/reset-password)",
			Handler:     p.userHandler,
		},
		{
			Name:        "client",
			Description: "Manage Keycloak clients",
			Handler:     p.clientHandler,
		},
		{
			Name:        "ssl",
			Description: "Configure SSL certificates",
			Handler:     p.sslHandler,
		},
		{
			Name:        "backup",
			Description: "Backup Keycloak configuration and data",
			Handler:     p.backupHandler,
		},
		{
			Name:        "restore",
			Description: "Restore Keycloak from backup",
			Handler:     p.restoreHandler,
		},
		{
			Name:        "configure",
			Description: "Interactive configuration management",
			Handler:     p.configureHandler,
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

func (p *Plugin) serviceActionHandler(action string) plugin.CommandHandler {
	return func(ctx context.Context, conn plugin.Connection, args []string, flags map[string]interface{}) error {
		_ = getSudoPass(flags) // For consistency with other handlers

		fmt.Printf("⚙️  %sing Keycloak services...\n", strings.Title(action))

		keycloakDir := "/opt/keycloak"
		dockerComposeCmd := fmt.Sprintf("cd %s && docker-compose %s", keycloakDir, action)

		if result := conn.RunCommand(dockerComposeCmd, plugin.WithHideOutput()); !result.Success {
			return fmt.Errorf("failed to %s Keycloak services: %s", action, result.Stderr)
		}

		fmt.Printf("✅ Keycloak services %sed successfully\n", action+"ed")
		return nil
	}
}

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

const dockerComposeTemplate = `version: '3.8'

services:
  keycloak-db:
    image: postgres:15
    container_name: keycloak-db
    environment:
      POSTGRES_DB: keycloak
      POSTGRES_USER: keycloak
      POSTGRES_PASSWORD: %s
    volumes:
      - keycloak_db_data:/var/lib/postgresql/data
    networks:
      - keycloak-network
    restart: unless-stopped

  keycloak:
    image: quay.io/keycloak/keycloak:23.0.0
    container_name: keycloak
    command: ["start-dev"]
    environment:
      KC_DB: postgres
      KC_DB_URL_HOST: keycloak-db
      KC_DB_URL_DATABASE: keycloak
      KC_DB_USERNAME: keycloak
      KC_DB_PASSWORD: %s
      KEYCLOAK_ADMIN: admin
      KEYCLOAK_ADMIN_PASSWORD: %s
      KC_HOSTNAME: %s
      KC_HTTP_ENABLED: true
      KC_HOSTNAME_STRICT: false
      KC_HOSTNAME_STRICT_HTTPS: false
    ports:
      - "8080:8080"
    depends_on:
      - keycloak-db
    networks:
      - keycloak-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health/ready"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  keycloak_db_data:

networks:
  keycloak-network:
    driver: bridge
`

const nginxTemplate = `server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`
