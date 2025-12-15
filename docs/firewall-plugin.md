# Firewall Plugin Documentation

## Overview

The firewall plugin provides comprehensive firewall management using UFW (Uncomplicated Firewall) for Ubuntu-based systems. It offers a simple interface for configuring, managing, and monitoring firewall rules.

## Features

- **UFW Installation**: Automatic installation and configuration of UFW
- **Rule Management**: Allow/deny traffic with flexible port and protocol options
- **Source IP Filtering**: Control access based on source IP addresses
- **Logging Configuration**: Configurable logging levels and verbosity
- **Safety Features**: SSH protection to prevent accidental lockouts
- **Rule Numbering**: Easy rule deletion with numbered rule listing
- **Status Monitoring**: Detailed firewall status and rule inspection

## Installation

```bash
vps-init user@server firewall install
```

### Installation Options

| Option | Shorthand | Default | Description |
|--------|-----------|---------|-------------|
| `--default-policy` | `-p` | `deny` | Default firewall policy (allow/deny) |
| `--enable-logging` | `-l` | `true` | Enable firewall logging |
| `--allow-ssh` | | `true` | Automatically allow SSH connections |

### Examples

```bash
# Install with default settings (deny incoming, allow SSH)
vps-init user@server firewall install

# Install with allow default policy
vps-init user@server firewall install --default-policy allow

# Install without logging
vps-init user@server firewall install --enable-logging=false

# Install without automatic SSH allowance
vps-init user@server firewall install --allow-ssh=false
```

## Usage

### Allow Traffic

```bash
# Allow a specific port
vps-init user@server firewall allow 80

# Allow a port with protocol
vps-init user@server firewall allow 443 tcp

# Allow from specific IP
vps-init user@server firewall allow 22 tcp 192.168.1.100

# Allow service names
vps-init user@server firewall allow ssh
vps-init user@server firewall allow http
vps-init user@server firewall allow https
```

### Deny Traffic

```bash
# Deny a specific port
vps-init user@server firewall deny 23

# Deny from specific IP
vps-init user@server firewall deny 22 tcp 192.168.1.50

# Deny all traffic from IP range
vps-init user@server firewall deny 192.168.1.0/24
```

### Firewall Management

```bash
# Enable firewall (activates all rules)
vps-init user@server firewall enable

# Disable firewall (deactivates all rules)
vps-init user@server firewall disable

# Show firewall status and rules
vps-init user@server firewall status

# Reset firewall to defaults
vps-init user@server firewall reset
```

### Rule Management

```bash
# Show numbered rules for deletion
vps-init user@server firewall status

# Delete specific rule by number
vps-init user@server firewall delete 3

# Configure logging
vps-init user@server firewall logging on
vps-init user@server firewall logging high
vps-init user@server firewall logging off
```

## Logging Levels

UFW supports different logging levels:

- `on`: Standard logging
- `off`: No logging
- `low`: Minimal logging
- `medium`: Moderate logging
- `high`: Verbose logging
- `full`: Maximum logging

## Common Use Cases

### Basic Web Server Setup

```bash
# Install and configure firewall
vps-init user@server firewall install

# Allow HTTP and HTTPS traffic
vps-init user@server firewall allow http
vps-init user@server firewall allow https

# Enable firewall
vps-init user@server firewall enable
```

### Database Server Access

```bash
# Allow MySQL only from specific IP
vps-init user@server firewall allow 3306 tcp 192.168.1.100

# Allow PostgreSQL from application server
vps-init user@server firewall allow 5432 tcp 10.0.1.50
```

### SSH Security

```bash
# Restrict SSH to specific IP ranges
vps-init user@server firewall delete 1  # Remove default SSH rule
vps-init user@server firewall allow 22 tcp 192.168.1.0/24
vps-init user@server firewall allow 22 tcp 10.0.0.0/8
```

### Development Environment

```bash
# Install with allow policy (more permissive for development)
vps-init user@server firewall install --default-policy allow

# Add specific restrictions as needed
vps-init user@server firewall deny 23    # Block telnet
vps-init user@server firewall deny 3389  # Block RDP
```

## Safety Features

### SSH Protection

The plugin automatically includes SSH protection to prevent accidental lockouts:

1. **Installation**: By default allows SSH connections unless explicitly disabled
2. **Enable Check**: Verifies SSH rule exists before enabling firewall
3. **Warning Messages**: Clear warnings about potential lockout scenarios

### Rule Validation

- Validates rule syntax before application
- Provides clear error messages for invalid configurations
- Shows numbered rules for easy management

### Logging Integration

- Configurable logging levels for security monitoring
- Detailed status output for troubleshooting
- Integration with system logging

## Advanced Configuration

### Custom Policies

```bash
# More restrictive installation
vps-init user@server firewall install \
  --default-policy deny \
  --enable-logging=true \
  --allow-ssh=false

# Then manually add specific rules
vps-init user@server firewall allow 22 tcp 192.168.1.0/24
vps-init user@server firewall allow 80
vps-init user@server firewall allow 443
```

### Service-Specific Rules

```bash
# Web server rules
vps-init user@server firewall allow 80    # HTTP
vps-init user@server firewall allow 443   # HTTPS
vps-init user@server firewall allow 8080  # Alternative HTTP

# Database rules
vps-init user@server firewall allow 3306  # MySQL
vps-init user@server firewall allow 5432  # PostgreSQL
vps-init user@server firewall allow 6379  # Redis

# Development tools
vps-init user@server firewall allow 3000  # Node.js apps
vps-init user@server firewall allow 8080  # Java apps
vps-init user@server firewall allow 9000  # Development servers
```

## Troubleshooting

### Common Issues

1. **Can't SSH after enabling firewall**
   ```bash
   # Check SSH rules
   vps-init user@server firewall status

   # Add SSH rule if missing
   vps-init user@server firewall allow ssh

   # Or allow from your specific IP
   vps-init user@server firewall allow 22 tcp <your-ip>
   ```

2. **Service not accessible**
   ```bash
   # Check firewall status
   vps-init user@server firewall status

   # Verify rule exists
   vps-init user@server firewall status | grep <port>

   # Add rule if missing
   vps-init user@server firewall allow <port>
   ```

3. **Resetting configuration**
   ```bash
   # Warning: This removes all rules
   vps-init user@server firewall reset

   # Reconfigure with install
   vps-init user@server firewall install
   ```

### Debugging

```bash
# Verbose status
vps-init user@server firewall status

# Check numbered rules
vps-init user@server firewall status | grep "\[.*\]"

# Enable logging for debugging
vps-init user@server firewall logging full

# Check system logs
ssh user@server "tail -f /var/log/ufw.log"
ssh user@server "sudo ufw status verbose"
```

## Security Best Practices

1. **Default to Deny**: Use `--default-policy deny` for better security
2. **Specific IP Ranges**: Restrict SSH to specific IP ranges when possible
3. **Enable Logging**: Use logging to monitor firewall activity
4. **Regular Reviews**: Periodically review firewall rules
5. **Test Changes**: Always test firewall changes in non-production environments

## Integration with Other Plugins

The firewall plugin works well with other VPS-Init plugins:

```bash
# Install firewall with fail2ban
vps-init user@server firewall install
vps-init user@server fail2ban install

# Secure web server setup
vps-init user@server firewall install
vps-init user@server nginx install
vps-init user@server firewall allow http
vps-init user@server firewall allow https
vps-init user@server firewall enable
```

## Plugin Metadata

- **Name**: firewall
- **Version**: 1.0.0
- **Author**: VPS-Init Team
- **License**: MIT
- **Repository**: github.com/wasilwamark/vps-init-plugins/firewall
- **Tags**: security, networking, firewall, ufw
- **Dependencies**: system (>=1.0.0)
- **Platforms**: linux/amd64, linux/arm64

## Contributing

To contribute to the firewall plugin:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## Support

For issues, questions, or contributions:
- GitHub Issues: [Repository Issues]
- Documentation: [VPS-Init Documentation]
- Community: [Discord/Slack Channel]