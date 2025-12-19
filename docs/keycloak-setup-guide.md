# Keycloak Setup Guide

This guide provides step-by-step instructions for setting up Keycloak identity and access management using VPS-Init.

## Prerequisites

Before installing Keycloak, ensure you have:

1. **VPS-Init Installed**: Follow the main README installation instructions
2. **Server Access**: SSH access to an Ubuntu/Debian server
3. **Domain Name**: (Recommended for production) A domain name pointing to your server
4. **Basic Dependencies**: Docker and Nginx (will be installed automatically)

## Quick Start

### 1. Add Server Alias

```bash
# Add your server with sudo password
vps-init alias add myserver user@your-server-ip --sudo-password 'your-sudo-password'
```

### 2. Install Dependencies

```bash
# Install Docker and Docker Compose
vps-init myserver docker install

# Install Nginx
vps-init myserver nginx install
```

### 3. Install Keycloak

```bash
# Basic installation (uses keycloak.local)
vps-init myserver keycloak install

# Or with custom domain
vps-init myserver keycloak install sso.yourdomain.com
```

### 4. Configure SSL (Production)

```bash
# Install SSL certificate
vps-init myserver keycloak ssl sso.yourdomain.com
```

### 5. Access Keycloak

- **Admin Console**: `http://sso.yourdomain.com/admin`
- **Base URL**: `http://sso.yourdomain.com`
- **Credentials**: Check `/opt/keycloak/credentials.txt` on your server

## Detailed Installation Guide

### Step 1: Server Preparation

Ensure your server meets the requirements:

```bash
# Check system resources (minimum 2GB RAM recommended)
vps-init myserver system status

# Update system packages
vps-init myserver system update
```

### Step 2: Install Dependencies

VPS-Init will handle all dependencies automatically:

```bash
# Install Docker Engine and Docker Compose
vps-init myserver docker install

# Verify Docker installation
vps-init myserver docker status

# Install Nginx for reverse proxy
vps-init myserver nginx install

# Verify Nginx installation
vps-init myserver nginx status
```

### Step 3: Install Keycloak

#### Option A: Basic Installation

```bash
# Install with default domain (keycloak.local)
vps-init myserver keycloak install
```

This creates:
- Keycloak server on port 8080 (internal)
- PostgreSQL database for persistence
- Nginx reverse proxy configuration
- Installation directory: `/opt/keycloak`

#### Option B: Custom Domain Installation

```bash
# Install with your domain
vps-init myserver keycloak install sso.yourdomain.com
```

### Step 4: Verify Installation

```bash
# Check Keycloak service status
vps-init myserver keycloak status

# View service logs
vps-init myserver keycloak logs

# Check Docker containers
vps-init myserver docker ps
```

### Step 5: Configure DNS

Point your domain to your server's IP address:

```
A    sso.yourdomain.com    YOUR_SERVER_IP
```

### Step 6: Configure SSL (Production Required)

```bash
# Install SSL certificate using Let's Encrypt
vps-init myserver keycloak ssl sso.yourdomain.com
```

This will:
- Install Certbot for SSL certificates
- Obtain SSL certificate for your domain
- Configure Nginx for HTTPS
- Set up HTTP to HTTPS redirect
- Update Keycloak configuration

### Step 7: First Login

1. Access the admin console: `https://sso.yourdomain.com/admin`
2. Use the credentials from `/opt/keycloak/credentials.txt`
3. You'll be prompted to change the admin password on first login

## Configuration

### Access Credentials

After installation, find your credentials:

```bash
# View credentials file
ssh your-server "cat /opt/keycloak/credentials.txt"
```

The file contains:
- Admin username: `admin`
- Admin password: Auto-generated secure password
- Database password: Auto-generated secure password

### Change Admin Password

#### Method 1: Through CLI

```bash
# Interactive password change
vps-init myserver keycloak configure

# Or directly using Keycloak admin CLI
ssh your-server "cd /opt/keycloak && docker-compose exec keycloak /opt/keycloak/bin/kcadm.sh update users/$(docker-compose exec -T keycloak /opt/keycloak/bin/kcadm.sh get users -r master -q username=admin --fields id --config /opt/keycloak/conf/keycloak-cli.properties | grep -o '\"id\":\"[^\"]*\"' | cut -d'\"' -f4) -r master -s 'credentials=[{\"type\":\"password\",\"value\":\"NEW_PASSWORD\",\"temporary\":false}]' --config /opt/keycloak/conf/keycloak-cli.properties"
```

#### Method 2: Through Admin Console

1. Log in to admin console
2. Navigate to `Master` realm → `Users` → `admin`
3. Click `Credentials` tab
4. Set new password

### Service Management

```bash
# Start Keycloak services
vps-init myserver keycloak start

# Stop Keycloak services
vps-init myserver keycloak stop

# Restart Keycloak services
vps-init myserver keycloak restart

# Check service health
vps-init myserver keycloak status

# View logs
vps-init myserver keycloak logs
vps-init myserver keycloak logs keycloak-db
```

## Production Setup

### Security Configuration

1. **Change Default Passwords**: Update admin and database passwords
2. **Enable SSL**: Always use HTTPS in production
3. **Firewall Rules**: Restrict access to necessary ports only
4. **Regular Updates**: Keep Keycloak and dependencies updated

### Performance Tuning

```bash
# Edit Docker Compose configuration
ssh your-server "nano /opt/keycloak/docker-compose.yml"

# Example production configuration:
environment:
  KC_CACHE: "local"
  KC_PROXY: "edge"
  KC_HOSTNAME_STRICT: "true"
  KC_HOSTNAME_STRICT_HTTPS: "true"
  KC_LOG_LEVEL: "INFO"
deploy:
  resources:
    limits:
      memory: 2G
    reservations:
      memory: 1G
```

### Backup Configuration

```bash
# Create backup
vps-init myserver keycloak backup

# Setup automated backups (daily at 2 AM)
ssh your-server << 'EOF'
cat << 'SCRIPT' > /opt/keycloak/backup.sh
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/var/backups/keycloak"
mkdir -p $BACKUP_DIR
tar -czf "$BACKUP_DIR/keycloak_backup_$DATE.tar.gz" /opt/keycloak
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
EOF

chmod +x /opt/keycloak/backup.sh
echo "0 2 * * * /opt/keycloak/backup.sh" | crontab -
EOF
```

## Integration Examples

### Single Sign-On (SSO) Setup

1. **Create Application Realm**:
   ```bash
   vps-init myserver keycloak realm create my-apps
   ```

2. **Create OAuth Client**:
   ```bash
   vps-init myserver keycloak client create web-app my-apps
   ```

3. **Configure Your Application**:
   - Use the Keycloak admin console to set redirect URLs
   - Configure client settings (access type, validity, etc.)
   - Test authentication flow

### Database Integration

Keycloak can integrate with external databases:

```bash
# Edit configuration to use external database
ssh your-server "nano /opt/keycloak/docker-compose.yml"

# Update environment variables:
environment:
  KC_DB: "postgres"
  KC_DB_URL: "jdbc:postgresql://external-db:5432/keycloak"
  KC_DB_USERNAME: "keycloak"
  KC_DB_PASSWORD: "your_password"
```

### LDAP Integration

1. **Access Admin Console**: Go to `User Federation`
2. **Add LDAP Provider**: Configure your LDAP server settings
3. **Test Connection**: Verify LDAP connectivity
4. **Sync Users**: Import users from LDAP directory

## Troubleshooting

### Common Issues

#### Keycloak Won't Start

```bash
# Check service status
vps-init myserver keycloak status

# View logs for errors
vps-init myserver keycloak logs keycloak

# Check Docker containers
vps-init myserver docker ps -a

# Restart services
vps-init myserver keycloak restart
```

#### SSL Certificate Issues

```bash
# Check Nginx configuration
ssh your-server "nginx -t"

# Verify certificate
ssh your-server "certbot certificates"

# Check Nginx logs
ssh your-server "sudo tail -f /var/log/nginx/error.log"

# Reissue certificate if needed
ssh your-server "certbot renew --dry-run"
```

#### Database Connection Problems

```bash
# Check database container
vps-init myserver docker logs keycloak-db

# Test database connectivity
ssh your-server "docker exec -it keycloak-db psql -U keycloak -d keycloak -c 'SELECT 1'"

# Check database logs
vps-init myserver keycloak logs keycloak-db
```

#### Performance Issues

```bash
# Check resource usage
vps-init myserver keycloak status

# Monitor Docker stats
ssh your-server "docker stats"

# Check system resources
ssh your-server "free -h && df -h && top"

# Check application logs for errors
vps-init myserver keycloak logs
```

### Getting Help

For additional support:

1. **Check Logs**: Always check service logs first
2. **Verify Configuration**: Ensure all settings are correct
3. **Test Dependencies**: Verify Docker, Nginx, and PostgreSQL are working
4. **Community**: Refer to Keycloak official documentation
5. **Issues**: Report bugs or issues to VPS-Init repository

## Migration Guide

### From Existing Keycloak Installation

1. **Export Current Configuration**:
   ```bash
   # Export realms, clients, and users
   # Use Keycloak export functionality or API
   ```

2. **Install New Keycloak**:
   ```bash
   vps-init myserver keycloak install sso.yourdomain.com
   ```

3. **Import Configuration**:
   ```bash
   # Import previously exported configuration
   # Use Keycloak admin console or CLI
   ```

### Backup and Restore

```bash
# Create backup of old installation
vps-init oldserver keycloak backup

# Transfer backup to new server
scp user@oldserver:/var/backups/keycloak/keycloak_backup_*.tar.gz ./

# Restore on new server
vps-init newserver keycloak restore ./keycloak_backup_*.tar.gz
```

## Best Practices

### Security

1. **Use HTTPS**: Always enable SSL in production
2. **Strong Passwords**: Use complex admin passwords
3. **Regular Updates**: Keep Keycloak updated
4. **Firewall Rules**: Restrict access to necessary ports
5. **Monitor Logs**: Regularly check for suspicious activity

### Performance

1. **Resource Allocation**: Ensure adequate memory and CPU
2. **Caching**: Configure appropriate caching settings
3. **Database Optimization**: Use proper database configuration
4. **Load Balancing**: Consider load balancing for high traffic
5. **Monitoring**: Set up monitoring and alerting

### Maintenance

1. **Regular Backups**: Schedule automated backups
2. **Log Rotation**: Configure log rotation
3. **Health Checks**: Implement health monitoring
4. **Documentation**: Keep configuration documented
5. **Testing**: Test disaster recovery procedures

This setup guide should help you successfully deploy and manage Keycloak using VPS-Init. For more advanced configurations, refer to the Keycloak plugin documentation and official Keycloak documentation.