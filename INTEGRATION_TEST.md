# Integration Tests with Testcontainers

This document outlines the integration testing plan using Testcontainers to provision Docker containers for each supported Linux distribution and run plugins against them.

## Overview

- **Testcontainers**: Docker-based testing framework
- **Goal**: Test all plugins across all supported distros
- **Scope**: Integration tests that verify real plugin execution

---

## Phase 1: Infrastructure Setup

- [x] Add testcontainers-go dependency to go.mod
- [x] Create `tests/integration/` directory structure
- [x] Set up Docker daemon access for testcontainers
- [x] Configure test environment variables (SSH credentials, etc.)

---

## Phase 2: Distro Container Setup

### Debian Family

- [x] Create Ubuntu 22.04 container test fixture
- [ ] Create Ubuntu 24.04 container test fixture
- [x] Create Debian 12 (Bookworm) container test fixture
- [x] Add systemd support for Ubuntu/Debian containers
- [x] Set up SSH server in containers
- [x] Create fixture helper functions for Debian family

### RedHat Family

- [ ] Create Fedora 40 container test fixture
- [ ] Create Rocky Linux 9 container test fixture
- [ ] Add systemd support for RedHat containers
- [ ] Set up SSH server in containers
- [ ] Create fixture helper functions for RedHat family

### Arch Family

- [ ] Create Arch Linux container test fixture
- [ ] Add systemd support for Arch container
- [ ] Set up SSH server in container
- [ ] Create fixture helper functions for Arch family

### Alpine Family

- [ ] Create Alpine 3.19 container test fixture
- [ ] Add OpenRC support for Alpine container
- [ ] Set up SSH server in container
- [ ] Create fixture helper functions for Alpine family

---

## Phase 3: Test Helper Infrastructure

- [x] Create base integration test suite structure
- [x] Implement container lifecycle management (start/stop/cleanup)
- [x] Create SSH connection helper for test containers
- [x] Add assertion helpers for command execution results
- [x] Implement test fixture setup/teardown
- [ ] Add logging infrastructure for test debugging
- [ ] Create test config generation utilities
- [ ] Implement test result reporting

---

## Phase 4: Core Plugin Tests

### System Plugin Tests

- [ ] Test `system update` on Ubuntu 22.04
- [ ] Test `system update` on Fedora 40
- [ ] Test `system update` on Arch Linux
- [ ] Test `system update` on Alpine 3.19
- [ ] Test `system upgrade` on Ubuntu 22.04
- [ ] Test `system upgrade` on Fedora 40
- [ ] Test `system upgrade` on Arch Linux
- [ ] Test `system upgrade` on Alpine 3.19
- [ ] Test `system clean` on all distros
- [ ] Test `system autoremove` on Debian/RedHat families

### Docker Plugin Tests

- [ ] Test `docker install` on Ubuntu 22.04
- [ ] Test `docker install` on Fedora 40
- [ ] Test `docker install` on Arch Linux
- [ ] Test Docker Compose installation
- [ ] Verify Docker service running post-install
- [ ] Test Docker daemon configuration
- [ ] Test user group setup for Docker

---

## Phase 5: Service Plugin Tests

### Nginx Plugin Tests

- [ ] Test `nginx install` on Ubuntu 22.04
- [ ] Test `nginx install` on Fedora 40
- [ ] Test `nginx install` on Arch Linux
- [ ] Test `nginx install-ssl` with self-signed cert
- [ ] Verify Nginx service status
- [ ] Test Nginx configuration validation
- [ ] Test SSL certificate generation
- [ ] Test basic site deployment

### MySQL Plugin Tests

- [ ] Test `mysql install` on Ubuntu 22.04
- [ ] Test `mysql install` on Fedora 40
- [ ] Test `mysql install` on Arch Linux
- [ ] Test `mysql create-db` database creation
- [ ] Test `mysql create-user` user creation
- [ ] Test `mysql grant-permissions` permissions
- [ ] Verify MySQL/MariaDB service status
- [ ] Test MySQL secure installation
- [ ] Test remote access configuration

### Redis Plugin Tests

- [ ] Test `redis install` on Ubuntu 22.04
- [ ] Test `redis install` on Fedora 40
- [ ] Test `redis install` on Arch Linux
- [ ] Verify Redis service status
- [ ] Test Redis configuration
- [ ] Test Redis persistence setup
- [ ] Test Redis authentication
- [ ] Test Redis benchmark

---

## Phase 6: Security Plugin Tests

### Fail2Ban Plugin Tests

- [ ] Test `fail2ban install` on Ubuntu 22.04
- [ ] Test `fail2ban install` on Fedora 40
- [ ] Test `fail2ban install` on Arch Linux
- [ ] Verify Fail2Ban service status
- [ ] Test SSH jail configuration
- [ ] Test ban/unban functionality
- [ ] Test custom jail creation
- [ ] Test whitelist configuration

### Firewall Plugin Tests

- [ ] Test `firewall install` (UFW) on Ubuntu 22.04
- [ ] Test `firewall install` (firewalld) on Fedora 40
- [ ] Test `firewall install` (iptables) on Alpine 3.19
- [ ] Test `firewall allow` for common ports
- [ ] Test `firewall deny` rules
- [ ] Test `firewall enable/disable`
- [ ] Test firewall rule persistence
- [ ] Test firewall rule listing

---

## Phase 7: Advanced Plugin Tests

### WireGuard Plugin Tests

- [ ] Test `wireguard install` on Ubuntu 22.04
- [ ] Test `wireguard install` on Fedora 40
- [ ] Test `wireguard install` on Arch Linux
- [ ] Test WireGuard key generation
- [ ] Test WireGuard config creation
- [ ] Test WireGuard service management
- [ ] Test peer configuration
- [ ] Test VPN connectivity

### Restic Plugin Tests

- [ ] Test `restic install` on Ubuntu 22.04
- [ ] Test `restic install` on Fedora 40
- [ ] Test `restic install` on Arch Linux
- [ ] Test restic repository initialization
- [ ] Test backup configuration
- [ ] Test restore functionality
- [ ] Test scheduled backups
- [ ] Test backup retention policies

### WordPress Plugin Tests

- [ ] Test `wordpress install` on Ubuntu 22.04
- [ ] Test `wordpress install` on Fedora 40
- [ ] Test WordPress with Nginx integration
- [ ] Test WordPress database setup
- [ ] Test WordPress configuration
- [ ] Test SSL integration
- [ ] Test WordPress update mechanism

### Keycloak Plugin Tests

- [ ] Test `keycloak install` on Ubuntu 22.04
- [ ] Test `keycloak install` on Fedora 40
- [ ] Test Java runtime dependency
- [ ] Test Keycloak service startup
- [ ] Test Keycloak admin configuration
- [ ] Test SSL certificate setup
- [ ] Test database integration (PostgreSQL/MySQL)

---

## Phase 8: Runtime Plugin Tests

- [ ] Test Node.js installation on all distros
- [ ] Test Python installation on all distros
- [ ] Test Go installation on all distros
- [ ] Test PHP installation on all distros
- [ ] Test version management
- [ ] Test runtime switching
- [ ] Test package manager integration (npm, pip, etc.)

---

## Phase 9: Cross-Distro Compatibility Tests

- [ ] Test plugin compatibility matrix
- [ ] Verify apt-based distros behave consistently
- [ ] Verify dnf/yum-based distros behave consistently
- [ ] Verify pacman-based distros behave consistently
- [ ] Verify apk-based distros behave consistently
- [ ] Test distro-specific workarounds
- [ ] Validate package name mappings

---

## Phase 10: Error Handling Tests

- [ ] Test plugin failure on missing dependencies
- [ ] Test plugin failure on invalid configurations
- [ ] Test graceful degradation scenarios
- [ ] Test rollback mechanisms
- [ ] Test error message clarity
- [ ] Test partial failure recovery

---

## Phase 11: Performance Tests

- [ ] Measure plugin installation times per distro
- [ ] Test concurrent plugin execution
- [ ] Test memory usage during plugin execution
- [ ] Test disk space usage
- [ ] Test SSH connection overhead
- [ ] Benchmark container startup times

---

## Phase 12: Integration Test CI/CD

- [ ] Create GitHub Actions workflow for integration tests
- [ ] Set up Docker-in-Docker (DinD) runner
- [ ] Configure test matrix for all distros
- [ ] Set up test result artifacts collection
- [ ] Configure test failure notifications
- [ ] Add test coverage reporting
- [ ] Set up test result history tracking

---

## Phase 13: Documentation & Maintenance

- [ ] Document test container setup process
- [ ] Document fixture helper API
- [ ] Create test debugging guide
- [ ] Document test environment requirements
- [ ] Set up test data update procedures
- [ ] Create test maintenance checklist
- [ ] Document known issues and workarounds

---

## Test Structure

```
tests/
├── integration/
│   ├── fixtures/
│   │   ├── containers.go       # Container fixture helpers
│   │   ├── debian.go           # Debian family fixtures
│   │   ├── redhat.go           # RedHat family fixtures
│   │   ├── arch.go             # Arch family fixtures
│   │   └── alpine.go           # Alpine family fixtures
│   ├── helpers/
│   │   ├── ssh.go              # SSH connection helpers
│   │   ├── assertions.go       # Test assertions
│   │   └── config.go           # Test config generation
│   ├── suite/
│   │   ├── suite.go            # Base test suite
│   │   └── setup.go            # Test setup/teardown
│   ├── plugins/
│   │   ├── system_test.go
│   │   ├── docker_test.go
│   │   ├── nginx_test.go
│   │   ├── mysql_test.go
│   │   ├── redis_test.go
│   │   ├── fail2ban_test.go
│   │   ├── firewall_test.go
│   │   ├── wireguard_test.go
│   │   ├── restic_test.go
│   │   ├── wordpress_test.go
│   │   ├── keycloak_test.go
│   │   └── runtimes_test.go
│   └── integration_test.go     # Main test entry point
```

---

## Running Tests

```bash
# Run all integration tests
go test ./tests/integration/... -v

# Run specific distro tests
go test ./tests/integration/... -v -run Ubuntu

# Run specific plugin tests
go test ./tests/integration/plugins -v -run TestDocker

# Run with test output
go test ./tests/integration/... -v -tags integration
```

---

## Requirements

- Docker daemon running
- Go 1.23+
- testcontainers-go
- SSH client (for testing SSH connections)
- Sufficient disk space for container images (~5GB)

---

## Notes

- Tests will use ephemeral containers
- Each test creates a fresh container
- Containers are cleaned up after tests
- Test data should be minimal to reduce test time
- Some tests may require network access for package installation
