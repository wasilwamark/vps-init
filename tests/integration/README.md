# Integration Tests

This directory contains integration tests for vps-init using testcontainers-go to provision Docker containers for each supported Linux distribution.

## Directory Structure

```
tests/integration/
├── fixtures/           # Container fixtures for different distros
│   ├── containers.go    # Base fixture implementation
│   ├── debian.go        # Debian family (Ubuntu, Debian)
│   ├── redhat.go       # RedHat family (Fedora, Rocky)
│   ├── arch.go         # Arch family (Arch Linux)
│   └── alpine.go       # Alpine family (Alpine)
├── helpers/            # Test helper functions
│   ├── ssh.go          # SSH connection helpers
│   ├── assertions.go   # Test assertions
│   └── config.go       # Test config generation
├── suite/              # Test suite infrastructure
│   ├── suite.go        # Base test suite
│   └── setup.go        # Test setup/teardown
├── plugins/            # Plugin-specific tests
│   ├── system_test.go
│   ├── docker_test.go
│   ├── nginx_test.go
│   └── ...
└── integration_test.go  # Main test entry point
```

## Prerequisites

- Docker daemon running
- Go 1.23+
- testcontainers-go (automatically installed via `go mod download`)

## Running Tests

### Run all integration tests

```bash
go test ./tests/integration/... -v
```

### Run specific distro tests

```bash
go test ./tests/integration/... -v -run Ubuntu
go test ./tests/integration/... -v -run Debian
```

### Run specific plugin tests

```bash
go test ./tests/integration/plugins -v -run TestSystem
go test ./tests/integration/plugins -v -run TestDocker
```

### Run with verbose output

```bash
go test ./tests/integration/... -v -tags integration
```

## Current Status

### Completed (Phase 1)

- [x] Add testcontainers-go dependency to go.mod
- [x] Create tests/integration/ directory structure
- [x] Set up base test suite structure
- [x] Create container fixture helpers
- [x] Create SSH connection helpers

### In Progress (Phase 2)

- [ ] Create RedHat family fixtures (Fedora, Rocky)
- [ ] Create Arch family fixture
- [ ] Create Alpine family fixture

### Pending (Phase 3+)

- [ ] Add assertion helpers
- [ ] Create test config generation utilities
- [ ] Implement plugin tests for all services

## Testcontainers Setup

Each test creates an ephemeral Docker container with:

- SSH server installed and configured
- Root access with password authentication
- Appropriate init system (systemd, OpenRC, etc.)

Test fixtures handle:

- Container creation and configuration
- SSH access setup
- Container lifecycle (start/stop/cleanup)
- Port mapping for SSH access

## Test Fixture API

```go
// Create a new test suite
ts := suite.NewTestSuite(t)

// Create a distro fixture
fixture := fixtures.NewUbuntuFixture(t, "22.04")

// Setup container with SSH
err := ts.SetupTestContainer(fixture)
defer ts.Teardown()

// Connect via SSH
sshClient, err := helpers.NewSSHClient(ts.SSHConfig)

// Run commands
output, err := sshClient.RunCommand("cat /etc/os-release")
```

## Available Fixtures

### Debian Family

```go
// Ubuntu 22.04 (Jammy)
fixture := fixtures.NewUbuntuFixture(t, "22.04")

// Ubuntu 24.04 (Noble)
fixture := fixtures.NewUbuntuFixture(t, "24.04")

// Debian 12 (Bookworm)
fixture := fixtures.NewDebianFixture(t, "bookworm")
```

### RedHat Family (Coming Soon)

```go
// Fedora 40
fixture := fixtures.NewFedoraFixture(t, "40")

// Rocky Linux 9
fixture := fixtures.NewRockyFixture(t, "9")
```

### Arch Family (Coming Soon)

```go
// Arch Linux
fixture := fixtures.NewArchFixture(t)
```

### Alpine Family (Coming Soon)

```go
// Alpine 3.19
fixture := fixtures.NewAlpineFixture(t, "3.19")
```

## Adding New Tests

### 1. Create a fixture for a new distro

See `tests/integration/fixtures/` for examples. Implement the `ContainerFixture` interface:

```go
type MyDistroFixture struct {
    *BaseFixture
    Version       string
    TestContainer testcontainers.Container
}

func (f *MyDistroFixture) CreateContainer(ctx context.Context) (testcontainers.Container, error) {
    // Create and configure container
}

func (f *MyDistroFixture) GetSSHConfig(ctx context.Context) (*helpers.SSHConfig, error) {
    // Return SSH connection details
}
```

### 2. Add integration test

```go
func TestMyDistro(t *testing.T) {
    ts := suite.NewTestSuite(t)
    fixture := fixtures.NewMyDistroFixture(t, "version")

    if err := ts.SetupTestContainer(fixture); err != nil {
        t.Fatalf("Failed to setup container: %v", err)
    }
    defer ts.Teardown()

    // Your test logic here
}
```

### 3. Add plugin test

```go
func TestSystemPlugin(t *testing.T) {
    ts := suite.NewTestSuite(t)
    fixture := fixtures.NewUbuntuFixture(t, "22.04")

    if err := ts.SetupTestContainer(fixture); err != nil {
        t.Fatalf("Failed to setup container: %v", err)
    }
    defer ts.Teardown()

    sshClient, err := helpers.NewSSHClient(ts.SSHConfig)
    ts.AssertNoError(err, "Failed to create SSH client")
    defer sshClient.Close()

    // Test plugin functionality
    output, err := sshClient.RunCommand("apt-get update")
    ts.AssertNoError(err, "Failed to update packages")
    ts.AssertContains(output, "Reading package lists")
}
```

## Troubleshooting

### Docker not running

Ensure Docker daemon is running and accessible:

```bash
docker ps
```

### Permission denied

On Linux, you may need to add your user to the docker group:

```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Container startup timeout

Increase timeout in fixture setup:

```go
time.Sleep(10 * time.Second) // Increase wait time
```

### SSH connection refused

Verify SSH is running inside container:

```bash
# From test logs, check container logs
docker logs <container-id>
```

## CI/CD Integration

See `.github/workflows/integration-tests.yml` for GitHub Actions configuration.

Tests run on:

- Ubuntu 22.04
- Ubuntu 24.04
- Debian 12
- Fedora 40
- Rocky Linux 9
- Arch Linux
- Alpine 3.19

## Resources

- [testcontainers-go Documentation](https://golang.testcontainers.org/)
- [testcontainers-go Examples](https://golang.testcontainers.org/examples/)
- [vps-init Documentation](../../README.md)
