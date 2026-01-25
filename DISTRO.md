# Distribution-Specific Command Execution Plan

## Todo List

- [ ] Analyze codebase for hardcoded distro-specific commands (DONE)
- [ ] Determine target Linux distributions to support: Ubuntu/Debian, CentOS/RHEL (7/8/9), Fedora, Arch Linux, Alpine
- [ ] Create distribution detection mechanism (read /etc/os-release file)
- [ ] Design distro info struct with package manager, service manager, and config paths
- [ ] Create package manager abstraction layer with methods for:
  - [ ] Update package lists
  - [ ] Install packages
  - [ ] Remove packages
  - [ ] Upgrade packages
  - [ ] Search packages
- [ ] Implement package manager adapters:
  - [ ] APT adapter (Debian/Ubuntu)
  - [ ] DNF adapter (Fedora, RHEL 8+, CentOS 8+)
  - [ ] YUM adapter (RHEL 7, CentOS 7)
  - [ ] Pacman adapter (Arch Linux)
  - [ ] APK adapter (Alpine Linux)
- [ ] Implement service management abstraction layer:
  - [ ] systemctl for systemd-based systems
  - [ ] service command for init.d systems
  - [ ] rc-service for OpenRC (Alpine)
- [ ] Update SSH connection methods:
  - [ ] Replace hardcoded `apt-get install` in `InstallPackage()`
  - [ ] Add `GetDistro()` method to Connection interface
  - [ ] Update `IsUbuntu()`, `IsDebian()`, `IsCentOS()`, `IsRedHat()` to use /etc/os-release
- [ ] Update service plugins:
  - [ ] docker/plugin.go - already uses get.docker.com (distro-agnostic)
  - [ ] nginx/plugin.go - replace apt-get with package manager abstraction
  - [ ] mysql/plugin.go - replace apt-get with package manager abstraction
  - [ ] redis/plugin.go - replace apt-get/apt with package manager abstraction
  - [ ] fail2ban/plugin.go - replace apt-get with package manager abstraction
  - [ ] wireguard/plugin.go - replace apt-get with package manager abstraction
  - [ ] restic/plugin.go - replace apt-get with package manager abstraction
  - [ ] firewall/plugin.go - replace apt-get/ufw with distro-specific firewall tools
  - [ ] system/plugin.go - replace all apt-get commands with package manager abstraction
  - [ ] runtimes/plugin.go - replace Node.js, Java, PHP, .NET installation commands
  - [ ] wordpress/plugin.go - replace apt-get commands
  - [ ] keycloak/plugin.go - replace apt-get commands
- [ ] Add distro-specific configuration paths:
  - [ ] nginx: /etc/nginx vs /etc/nginx/nginx.conf location
  - [ ] mariadb/mysql: /etc/mysql vs /etc/my.cnf
  - [ ] redis: /etc/redis vs /etc/redis.conf
  - [ ] fail2ban: /etc/fail2ban vs /etc/fail2ban/jail.local
  - [ ] wireguard: /etc/wireguard (common)
- [ ] Handle distro-specific package names:
  - [ ] mariadb-server vs mysql-server
  - [ ] python3-certbot-nginx vs python3-certbot-apache
  - [ ] ufw vs firewalld vs iptables
  - [ ] openjdk-8-jdk vs java-1.8.0-openjdk-devel
- [ ] Create fallback mechanisms for unsupported distributions
- [ ] Add logging for detected distribution and used commands
- [ ] Test on each target distribution (or via container/VM)
- [ ] Update documentation with supported distributions and examples

## Current Hardcoded Commands Found

### SSH Connection (internal/ssh/ssh.go)

- Line 630: `apt-get install -y` in `InstallPackage()`
- Lines 636-656: Detection using `lsb_release -si`

### System Plugin (internal/services/system/plugin.go)

- Lines 154-161: `apt-get update`, `apt-get install -y`
- Lines 170, 185, 198, 211, 235, 256: All apt-get commands

### Nginx Plugin (internal/services/nginx/plugin.go)

- Lines 154, 159: `apt-get update`, `apt-get install -y nginx`
- Lines 447-448: `apt-get install -y certbot python3-certbot-nginx`

### MySQL Plugin (internal/services/mysql/plugin.go)

- Lines 97, 102: `apt-get update`, `apt-get install -y mariadb-server`

### Redis Plugin (internal/services/redis/plugin.go)

- Lines 159, 165: `apt update`, `apt install -y redis-server`
- Lines 199: `apt remove --purge -y`

### Fail2Ban Plugin (internal/services/fail2ban/plugin.go)

- Lines 117, 120: `apt-get update`, `apt-get install -y fail2ban`

### WireGuard Plugin (internal/services/wireguard/plugin.go)

- Lines 109, 115: `apt-get update`, `apt-get install -y wireguard wireguard-tools qrencode iptables`

### Restic Plugin (internal/services/restic/plugin.go)

- Lines 104, 109: `apt-get update`, `apt-get install -y restic`

### Firewall Plugin (internal/services/firewall/plugin.go)

- Lines 231, 237: `apt update`, `apt install -y ufw`
- Uses UFW which is Debian/Ubuntu specific

### Runtimes Plugin (internal/services/runtimes/plugin.go)

- Lines 211, 308-314, 378-396, 435, 485, 495, 708, 726: Multiple apt-get commands for Node.js, Java, PHP, .NET

### WordPress Plugin (internal/services/wordpress/plugin.go)

- Lines 94, 98: `apt-get update`, `apt-get install -y`

### Keycloak Plugin (internal/services/keycloak/keycloak.go)

- Lines 433-434: `apt-get update`, `apt-get install -y certbot python3-certbot-nginx`

### Docker Plugin (internal/services/docker/plugin.go)

- Uses get.docker.com convenience script (handles distro differences automatically)
