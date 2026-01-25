package fixtures

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type BaseFixture struct {
	T            *testing.T
	ImageName    string
	ExposedPorts []string
	EnvVars      map[string]string
}

func NewBaseFixture(t *testing.T, imageName string) *BaseFixture {
	return &BaseFixture{
		T:            t,
		ImageName:    imageName,
		ExposedPorts: []string{"22/tcp"},
		EnvVars:      make(map[string]string),
	}
}

func (f *BaseFixture) WithExposedPorts(ports ...string) *BaseFixture {
	f.ExposedPorts = append(f.ExposedPorts, ports...)
	return f
}

func (f *BaseFixture) WithEnv(key, value string) *BaseFixture {
	f.EnvVars[key] = value
	return f
}

func (f *BaseFixture) CreateContainerRequest() *testcontainers.GenericContainerRequest {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        f.ImageName,
			ExposedPorts: f.ExposedPorts,
			Env:          f.EnvVars,
			WaitingFor:   wait.ForLog("sshd"),
		},
		Started: false,
	}
	return &req
}
