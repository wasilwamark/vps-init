package integration

import (
	"testing"

	"github.com/wasilwamark/vps-init/tests/integration/fixtures"
	"github.com/wasilwamark/vps-init/tests/integration/helpers"
	"github.com/wasilwamark/vps-init/tests/integration/suite"
)

func TestUbuntu2204Container(t *testing.T) {
	ts := suite.NewTestSuite(t)
	fixture := fixtures.NewUbuntuFixture(t, "22.04")

	if err := ts.SetupTestContainer(fixture); err != nil {
		t.Fatalf("Failed to setup container: %v", err)
	}
	defer ts.Teardown()

	sshConfig := &helpers.SSHConfig{
		Host:     ts.SSHConfig.Host,
		Port:     ts.SSHConfig.Port,
		User:     ts.SSHConfig.User,
		Password: ts.SSHConfig.Password,
	}

	sshClient, err := helpers.NewSSHClient(sshConfig)
	ts.AssertNoError(err, "Failed to create SSH client")
	defer sshClient.Close()

	err = sshClient.TestConnection()
	ts.AssertNoError(err, "Failed to test SSH connection")

	output, err := sshClient.RunCommand("cat /etc/os-release | grep PRETTY_NAME")
	ts.AssertNoError(err, "Failed to run command")
	ts.AssertContains(output, "Ubuntu 22.04")
}

func TestDebian12Container(t *testing.T) {
	ts := suite.NewTestSuite(t)
	fixture := fixtures.NewDebianFixture(t, "bookworm")

	if err := ts.SetupTestContainer(fixture); err != nil {
		t.Fatalf("Failed to setup container: %v", err)
	}
	defer ts.Teardown()

	sshConfig := &helpers.SSHConfig{
		Host:     ts.SSHConfig.Host,
		Port:     ts.SSHConfig.Port,
		User:     ts.SSHConfig.User,
		Password: ts.SSHConfig.Password,
	}

	sshClient, err := helpers.NewSSHClient(sshConfig)
	ts.AssertNoError(err, "Failed to create SSH client")
	defer sshClient.Close()

	err = sshClient.TestConnection()
	ts.AssertNoError(err, "Failed to test SSH connection")

	output, err := sshClient.RunCommand("cat /etc/os-release | grep PRETTY_NAME")
	ts.AssertNoError(err, "Failed to run command")
	ts.AssertContains(output, "Debian GNU/Linux")
}
