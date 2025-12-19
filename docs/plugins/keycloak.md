# Keycloak Plugin

## Overview

The Keycloak plugin provides comprehensive identity and access management (IAM) capabilities using Keycloak, an open-source IAM solution. It deploys Keycloak with PostgreSQL database, automatic SSL configuration, and includes tools for managing realms, users, and clients.

## Features

- **Docker-based Deployment**: Complete Keycloak + PostgreSQL stack using Docker Compose
- **Automatic SSL/TLS**: Let's Encrypt integration with Nginx reverse proxy
- **Service Management**: Start, stop, restart, and monitor Keycloak services
- **Realm Management**: Create, list, and delete Keycloak realms
- **User Administration**: Manage users, reset passwords, and create accounts
- **Client Configuration**: OAuth/OpenID Connect client management
- **Backup & Restore**: Complete data backup and recovery
- **Health Monitoring**: Service health checks and status reporting
- **Interactive Configuration**: User-friendly setup and management

## Prerequisites

Before using the Keycloak plugin, ensure the following dependencies are installed:

- **Docker**: Container runtime for Keycloak services
- **Docker Compose**: Orchestration tool for multi-container applications
- **Nginx**: Web server for reverse proxy configuration
- **Domain Name**: (Recommended for SSL) A domain name pointing to your server

You can install dependencies using:
```bash
vps-init user@server docker install
vps-init user@server nginx install
```

## Installation

### Basic Installation

Deploy Keycloak with default configuration:

```bash
vps-init user@server keycloak install
```

### Custom Domain Installation

Deploy Keycloak with a custom domain name:

```bash
vps-init user@server keycloak install keycloak.example.com
```

### What Gets Installed

- **Keycloak Server**: Version 23.0.0 in development mode
- **PostgreSQL Database**: Version 15 for data persistence
- **Nginx Reverse Proxy**: HTTP proxy with SSL support
- **Docker Network**: Isolated network for services
- **Persistent Storage**: Database volumes for data retention

## Configuration

### Access Information

After installation, you'll receive:

- **Admin Console**: `http://your-domain/admin`
- **Base URL**: `http://your-domain`
- **Admin Credentials**: Saved to `/opt/keycloak/credentials.txt`
- **Installation Directory**: `/opt/keycloak`

### SSL Configuration

Enable HTTPS with Let's Encrypt:

```bash
vps-init user@server keycloak ssl your-domain.com
```

This will:
- Install and configure Certbot
- Obtain SSL certificate
- Update Nginx configuration for HTTPS
- Enable HTTP to HTTPS redirect
- Restart Keycloak with SSL settings

## Commands

### Service Management

```bash
# Start Keycloak services
vps-init user@server keycloak start

# Stop Keycloak services
vps-init user@server keycloak stop

# Restart Keycloak services
vps-init user@server keycloak restart

# Check service status
vps-init user@server keycloak status

# View service logs
vps-init user@server keycloak logs [keycloak|keycloak-db]
```

### Realm Management

```bash
# List all realms
vps-init user@server keycloak realm list

# Create a new realm
vps-init user@server keycloak realm create my-realm

# Delete a realm
vps-init user@server keycloak realm delete my-realm
```

### User Management

```bash
# List users in default (master) realm
vps-init user@server keycloak user list

# Create a new user
vps-init user@server keycloak user create john.doe

# Create user in specific realm
vps-init user@server keycloak user create john.doe my-realm

# Reset user password
vps-init user@server keycloak user reset-password john.doe
```

### Client Management

```bash
# List all clients
vps-init user@server keycloak client list

# Create a new client
vps-init user@server keycloak client create my-app

# Create client in specific realm
vps-init user@server keycloak client create my-app my-realm
```

### Backup and Restore

```bash
# Create backup
vps-init user@server keycloak backup

# Restore from backup
vps-init user@server keycloak restore /var/backups/keycloak/keycloak_backup_20241219_120000.tar.gz
```

### Interactive Configuration

Access the interactive configuration menu:

```bash
vps-init user@server keycloak configure
```

This menu provides options to:
- View current configuration
- Update admin password
- Change domain settings
- Check service status
- Display access URLs

### Uninstallation

Remove Keycloak completely:

```bash
vps-init user@server keycloak uninstall [domain]
```

This will:
- Stop and remove all containers
- Remove Nginx configuration
- Delete installation directory
- Clean up Docker images and volumes

## Security Considerations

### Password Management

- Admin and database passwords are randomly generated
- Credentials are stored in `/opt/keycloak/credentials.txt` with secure permissions (600)
- Consider using a password manager for credential storage

### SSL/TLS

- Always use HTTPS in production environments
- Configure SSL after initial installation
- Set up automatic certificate renewal

### Network Security

- Keycloak is accessible through Nginx reverse proxy only
- Docker containers run in isolated network
- Consider firewall rules for additional protection

## Troubleshooting

### Service Won't Start

```bash
# Check service status
vps-init user@server keycloak status

# View logs for errors
vps-init user@server keycloak logs keycloak

# Check Docker containers
ssh user@server "docker ps -a"
```

### SSL Certificate Issues

```bash
# Check Nginx configuration
ssh user@server "nginx -t"

# View SSL certificate status
ssh user@server "certbot certificates"

# Check Nginx logs
ssh user@server "sudo tail -f /var/log/nginx/error.log"
```

### Database Connection Problems

```bash
# Check database container
ssh user@server "docker logs keycloak-db"

# Test database connectivity
ssh user@server "docker exec -it keycloak-db psql -U keycloak -d keycloak -c 'SELECT 1'"
```

### Performance Issues

```bash
# Check resource usage
vps-init user@server keycloak status

# Monitor Docker stats
ssh user@server "docker stats"

# Check system resources
ssh user@server "free -h && df -h"
```

## Integration Examples

### Single Sign-On (SSO) Setup

1. **Create Application Realm**:
   ```bash
   vps-init user@server keycloak realm create my-apps
   ```

2. **Create OAuth Client**:
   ```bash
   vps-init user@server keycloak client create web-app my-apps
   ```

3. **Create User**:
   ```bash
   vps-init user@server keycloak user create app-user my-apps
   ```

4. **Configure Your Application**:
   Use the Keycloak admin console to configure redirect URLs and client settings.

### Backup Automation

```bash
# Create backup script on server
cat << 'EOF' > /opt/keycloak/backup.sh
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/var/backups/keycloak"
mkdir -p $BACKUP_DIR
tar -czf "$BACKUP_DIR/keycloak_backup_$DATE.tar.gz" /opt/keycloak
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
EOF

chmod +x /opt/keycloak/backup.sh

# Add to cron for daily backups
echo "0 2 * * * /opt/keycloak/backup.sh" | crontab -
```

## Advanced Configuration

### Custom Docker Compose

You can modify the Docker Compose configuration at `/opt/keycloak/docker-compose.yml`:

```yaml
# Example: Production-ready configuration
services:
  keycloak:
    environment:
      KC_CACHE: "local"
      KC_DB: "postgres"
      KC_DB_URL_HOST: "keycloak-db"
      KC_DB_URL_DATABASE: "keycloak"
      KC_DB_USERNAME: "keycloak"
      KC_DB_PASSWORD: "${DB_PASSWORD}"
      KEYCLOAK_ADMIN: "admin"
      KEYCLOAK_ADMIN_PASSWORD: "${ADMIN_PASSWORD}"
      KC_HOSTNAME: "keycloak.example.com"
      KC_HTTP_ENABLED: "true"
      KC_PROXY: "edge"
    ports:
      - "8080:8080"
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M
```

### Environment Variables

Keycloak can be configured using environment variables in the Docker Compose file:

- `KC_CACHE`: Cache configuration (local, redis, etc.)
- `KC_PROXY`: Proxy mode (none, edge, reencrypt)
- `KC_HTTP_ENABLED`: Enable HTTP endpoint
- `KC_HOSTNAME_STRICT`: Enforce hostname
- `KC_LOG_LEVEL`: Logging level (INFO, DEBUG, WARN)

## API Access

Keycloak provides REST APIs for programmatic access:

- **Admin API**: `http://your-domain/admin/realms`
- **OpenID Configuration**: `http://your-domain/realms/master/.well-known/openid_configuration`
- **Token Endpoint**: `http://your-domain/realms/master/protocol/openid-connect/token`

## Support

For issues related to the Keycloak plugin, check:

1. Service logs: `vps-init user@server keycloak logs`
2. Service status: `vps-init user@server keycloak status`
3. Keycloak documentation: https://www.keycloak.org/documentation

For Keycloak-specific questions and advanced configuration, refer to the official Keycloak documentation.