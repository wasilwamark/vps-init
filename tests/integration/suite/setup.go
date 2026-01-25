package suite

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/wasilwamark/vps-init/tests/integration/helpers"
)

func (s *TestSuite) SetupTestContainer(fixture ContainerFixture) error {
	ctx := context.Background()

	container, err := fixture.CreateContainer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	s.TestContainer = container

	if err := container.Start(ctx); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	time.Sleep(2 * time.Second)

	sshConfig, err := fixture.GetSSHConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get SSH config: %w", err)
	}

	s.SSHConfig = &helpers.SSHConfig{
		Host:     sshConfig.Host,
		Port:     sshConfig.Port,
		User:     sshConfig.User,
		Password: sshConfig.Password,
		KeyPath:  sshConfig.KeyPath,
	}

	s.Cleanup = func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			s.T.Logf("failed to terminate container: %v", err)
		}
	}

	return nil
}

type ContainerFixture interface {
	CreateContainer(ctx context.Context) (testcontainers.Container, error)
	GetSSHConfig(ctx context.Context) (*helpers.SSHConfig, error)
}
