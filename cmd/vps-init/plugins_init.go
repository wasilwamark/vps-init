package main

import (
	"github.com/wasilwamark/vps-init/internal/core/alias"
	pluginmanager "github.com/wasilwamark/vps-init/internal/core/plugin-manager"

	"github.com/wasilwamark/vps-init/internal/services/docker"
	"github.com/wasilwamark/vps-init/internal/services/fail2ban"
	"github.com/wasilwamark/vps-init/internal/services/firewall"
	"github.com/wasilwamark/vps-init/internal/services/nginx"
	"github.com/wasilwamark/vps-init/internal/services/redis"
	"github.com/wasilwamark/vps-init/internal/services/runtimes"
	"github.com/wasilwamark/vps-init/internal/services/system"
	"github.com/wasilwamark/vps-init/internal/services/wireguard"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// initializeBuiltinPlugins registers all built-in plugins
func initializeBuiltinPlugins() {
	// Register core plugins
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/core/alias", alias.NewPlugin())
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/core/plugin-manager", pluginmanager.NewPlugin())

	// Register service plugins
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/system", &system.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/docker", &docker.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/fail2ban", &fail2ban.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/firewall", &firewall.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/nginx", &nginx.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/wireguard", &wireguard.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/redis", &redis.Plugin{})
	plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/runtimes", &runtimes.Plugin{})
	// TODO: Update remaining plugins to enhanced interface
	// plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/mysql", &mysql.Plugin{})
	// plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/wordpress", &wordpress.Plugin{})
	// plugin.RegisterBuiltin("github.com/wasilwamark/vps-init/services/restic", &restic.Plugin{})
}
