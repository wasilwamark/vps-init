# Keycloak Plugin Cheat Sheet

Quick reference for Keycloak plugin commands and common operations.

## Installation

```bash
# Basic installation
vps-init myserver keycloak install

# Custom domain
vps-init myserver keycloak install sso.example.com

# Install SSL
vps-init myserver keycloak ssl sso.example.com
```

## Service Management

```bash
# Start/stop/restart
vps-init myserver keycloak start
vps-init myserver keycloak stop
vps-init myserver keycloak restart

# Status and logs
vps-init myserver keycloak status
vps-init myserver keycloak logs
vps-init myserver keycloak logs keycloak-db
```

## Realm Management

```bash
# List realms
vps-init myserver keycloak realm list

# Create realm
vps-init myserver keycloak realm create my-realm

# Delete realm
vps-init myserver keycloak realm delete my-realm
```

## User Management

```bash
# List users (master realm)
vps-init myserver keycloak user list

# Create user
vps-init myserver keycloak user create username

# Create user in specific realm
vps-init myserver keycloak user create username my-realm

# Reset password
vps-init myserver keycloak user reset-password username
```

## Client Management

```bash
# List clients
vps-init myserver keycloak client list

# Create client
vps-init myserver keycloak client create client-name

# Create client in realm
vps-init myserver keycloak client create client-name my-realm
```

## Backup & Restore

```bash
# Create backup
vps-init myserver keycloak backup

# Restore from backup
vps-init myserver keycloak restore /path/to/backup.tar.gz

# Interactive configuration
vps-init myserver keycloak configure
```

## Common URLs

After installation:
- **Admin Console**: `https://your-domain/admin`
- **Base URL**: `https://your-domain`
- **API Docs**: `https://your-domain/realms/master/.well-known/openid_configuration`

## Files and Locations

- **Installation**: `/opt/keycloak`
- **Credentials**: `/opt/keycloak/credentials.txt`
- **Config**: `/opt/keycloak/docker-compose.yml`
- **Backups**: `/var/backups/keycloak/`
- **Nginx Config**: `/etc/nginx/sites-available/your-domain`

## Troubleshooting Commands

```bash
# Check all services
vps-init myserver keycloak status

# Check Docker containers
vps-init myserver docker ps

# Test HTTP response
curl -f http://your-domain/health/ready

# Test HTTPS response
curl -f https://your-domain/health/ready

# Check Nginx config
ssh myserver "nginx -t"

# View SSL certificates
ssh myserver "certbot certificates"
```

## Quick Setup Sequence

```bash
# 1. Install dependencies
vps-init myserver docker install
vps-init myserver nginx install

# 2. Install Keycloak
vps-init myserver keycloak install sso.example.com

# 3. Configure SSL
vps-init myserver keycloak ssl sso.example.com

# 4. Create realm for apps
vps-init myserver keycloak realm create my-apps

# 5. Create admin user for realm
vps-init myserver keycloak user create admin-user my-apps

# 6. Create OAuth client
vps-init myserver keycloak client create web-app my-apps

# 7. Create backup
vps-init myserver keycloak backup
```

## Environment Variables

Keycloak configuration can be modified in `/opt/keycloak/docker-compose.yml`:

```yaml
environment:
  KC_HOSTNAME: "sso.example.com"
  KC_PROXY: "edge"
  KC_CACHE: "local"
  KC_LOG_LEVEL: "INFO"
  KC_HOSTNAME_STRICT: "true"
  KC_HOSTNAME_STRICT_HTTPS: "true"
```

## Port Information

- **Keycloak Internal**: 8080 (container)
- **Nginx HTTP**: 80
- **Nginx HTTPS**: 443
- **PostgreSQL**: 5432 (internal to Docker network)

## Monitoring Commands

```bash
# Resource usage
vps-init myserver keycloak status

# Docker stats
ssh myserver "docker stats"

# System resources
ssh myserver "free -h && df -h"

# Recent logs
vps-init myserver keycloak logs | tail -50
```

## Security Checklist

- [ ] SSL certificate installed and valid
- [ ] Admin password changed from default
- [ ] Firewall rules configured
- [ ] Regular backups scheduled
- [ ] Monitoring and alerting set up
- [ ] Log rotation configured
- [ ] Software updates applied