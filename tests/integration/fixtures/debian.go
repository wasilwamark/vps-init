package fixtures

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wasilwamark/vps-init/tests/integration/helpers"
)

type UbuntuFixture struct {
	*BaseFixture
	Version       string
	TestContainer testcontainers.Container
}

func NewUbuntuFixture(t *testing.T, version string) *UbuntuFixture {
	return &UbuntuFixture{
		BaseFixture: NewBaseFixture(t, fmt.Sprintf("ubuntu:%s", version)),
		Version:     version,
	}
}

func (f *UbuntuFixture) CreateContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        f.ImageName,
			ExposedPorts: f.ExposedPorts,
			Env:          f.EnvVars,
			Cmd:          []string{"/bin/bash", "-c", "apt-get update && apt-get install -y openssh-server && mkdir -p /var/run/sshd && echo 'root:testpass' | chpasswd && sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config && /usr/sbin/sshd && tail -f /dev/null"},
			WaitingFor: wait.ForAll(
				wait.ForLog("sshd"),
				wait.ForListeningPort("22/tcp"),
			),
		},
		Started: false,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}
	f.TestContainer = container
	return container, nil
}

func (f *UbuntuFixture) GetSSHConfig(ctx context.Context) (*helpers.SSHConfig, error) {
	host, err := f.TestContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := f.TestContainer.MappedPort(ctx, "22/tcp")
	if err != nil {
		return nil, err
	}

	return &helpers.SSHConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "root",
		Password: "testpass",
	}, nil
}

type DebianFixture struct {
	*BaseFixture
	Version       string
	TestContainer testcontainers.Container
}

func NewDebianFixture(t *testing.T, version string) *DebianFixture {
	return &DebianFixture{
		BaseFixture: NewBaseFixture(t, fmt.Sprintf("debian:%s", version)),
		Version:     version,
	}
}

func (f *DebianFixture) CreateContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        f.ImageName,
			ExposedPorts: f.ExposedPorts,
			Env:          f.EnvVars,
			Cmd:          []string{"/bin/bash", "-c", "apt-get update && apt-get install -y openssh-server && mkdir -p /var/run/sshd && echo 'root:testpass' | chpasswd && sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config && /usr/sbin/sshd && tail -f /dev/null"},
			WaitingFor: wait.ForAll(
				wait.ForLog("sshd"),
				wait.ForListeningPort("22/tcp"),
			),
		},
		Started: false,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}
	f.TestContainer = container
	return container, nil
}

func (f *DebianFixture) GetSSHConfig(ctx context.Context) (*helpers.SSHConfig, error) {
	host, err := f.TestContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := f.TestContainer.MappedPort(ctx, "22/tcp")
	if err != nil {
		return nil, err
	}

	return &helpers.SSHConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "root",
		Password: "testpass",
	}, nil
}
