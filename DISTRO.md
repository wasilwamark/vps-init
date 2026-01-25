# Distribution-Specific Command Execution

## Implementation Status: ✅ COMPLETED

All tasks completed. VPS-Init now supports distribution-specific command execution for:

- Ubuntu/Debian (APT)
- CentOS/RHEL 7 (YUM)
- CentOS/RHEL 8+ & Fedora (DNF)
- Arch Linux (Pacman)
- Alpine Linux (APK)

## Completed Tasks

### Core Infrastructure

- ✅ Created distribution detection mechanism (reads /etc/os-release)
- ✅ Designed distro info struct with package manager, service manager fields
- ✅ Created package manager abstraction layer with methods:
  - ✅ Update package lists
  - ✅ Install packages
  - ✅ Remove packages
  - ✅ Upgrade packages
  - ✅ Search packages
- ✅ Implemented package manager adapters:
  - ✅ APT adapter (Debian/Ubuntu)
  - ✅ DNF adapter (Fedora, RHEL 8+, CentOS 8+)
  - ✅ YUM adapter (RHEL 7, CentOS 7)
  - ✅ Pacman adapter (Arch Linux)
  - ✅ APK adapter (Alpine Linux)
- ✅ Updated SSH connection methods:
  - ✅ Added GetDistroInfo() method to Connection interface
  - ✅ Updated InstallPackage() to use distro-aware commands
  - ✅ Updated IsUbuntu(), IsDebian(), IsCentOS(), IsRedHat() to use /etc/os-release

### Service Plugins Updated

- ✅ system/plugin.go - all package commands use package manager abstraction
- ✅ nginx/plugin.go - uses package manager abstraction
- ✅ mysql/plugin.go - uses package manager abstraction
- ✅ redis/plugin.go - uses package manager abstraction
- ✅ fail2ban/plugin.go - uses package manager abstraction
- ✅ wireguard/plugin.go - uses package manager abstraction
- ✅ restic/plugin.go - uses package manager abstraction
- ✅ firewall/plugin.go - uses package manager abstraction + distro-specific firewall detection (UFW for Debian, firewalld for RHEL)
- ✅ runtimes/plugin.go - added package manager helper
- ✅ wordpress/plugin.go - uses package manager abstraction
- ✅ keycloak/plugin.go - uses package manager abstraction
- ✅ Added logging for detected distribution and executed commands

## Logging

All operations now log:

- Detected distribution (name and version)
- Package manager being used
- Commands being executed

Example output:

```
ℹ️  Detected Distribution: Ubuntu 22.04
📦 Using Package Manager: apt
⚡ Executing: DEBIAN_FRONTEND=noninteractive apt-get install -y nginx
```

## Remaining Work (Future Enhancements)

### Package Name Mappings

Some packages may have different names across distributions:

- mariadb-server vs mysql-server
- python3-certbot-nginx vs python3-certbot-apache
- openjdk-8-jdk vs java-1.8.0-openjdk-devel

### Configuration Paths

Different distributions use different configuration paths:

- nginx: /etc/nginx vs /etc/nginx/nginx.conf location
- mariadb/mysql: /etc/mysql vs /etc/my.cnf
- redis: /etc/redis vs /etc/redis.conf
- fail2ban: /etc/fail2ban vs /etc/fail2ban/jail.local

### Testing

Test on each target distribution (via container/VM) to verify package commands work correctly.

### Docker Plugin (internal/services/docker/plugin.go)

- Uses get.docker.com convenience script (handles distro differences automatically)
