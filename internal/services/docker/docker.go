package docker

import (
	"fmt"
	"github.com/wasilwamark/vps-init/internal/ssh"
)

type Service struct {
	ssh *ssh.Connection
}

func New(ssh *ssh.Connection) *Service {
	return &Service{ssh: ssh}
}

func (s *Service) Install() bool {
	fmt.Println("🐳 Installing Docker...")

	// Install prerequisites
	commands := []string{
		"apt-get update",
		"apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release",
	}

	for _, cmd := range commands {
		result := s.ssh.RunCommand(cmd, false)
		if !result.Success {
			fmt.Printf("❌ Failed to run: %s\n", cmd)
			return false
		}
	}

	// Add Docker's official GPG key
	addKeyCmd := `curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg`
	result := s.ssh.RunCommand(addKeyCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to add Docker GPG key")
		return false
	}

	// Add Docker repository
	repoCmd := `echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list`
	result = s.ssh.RunCommand(repoCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to add Docker repository")
		return false
	}

	// Update and install Docker
	installCmd := "apt-get update && apt-get install -y docker-ce docker-ce-cli containerd.io"
	result = s.ssh.RunCommand(installCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to install Docker")
		return false
	}

	// Start and enable Docker
	if !s.ssh.Systemctl("start", "docker") {
		fmt.Println("❌ Failed to start Docker")
		return false
	}

	if !s.ssh.Systemctl("enable", "docker") {
		fmt.Println("❌ Failed to enable Docker")
		return false
	}

	// Add user to docker group
	userCmd := "usermod -aG docker ubuntu"
	result = s.ssh.RunCommand(userCmd, false)
	if !result.Success {
		fmt.Println("⚠️  Could not add ubuntu user to docker group (might not exist)")
	}

	// Install Docker Compose
	fmt.Println("📦 Installing Docker Compose...")
	composeCmd := `curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose`
	result = s.ssh.RunCommand(composeCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to download Docker Compose")
		return false
	}

	// Make docker-compose executable
	chmodCmd := "chmod +x /usr/local/bin/docker-compose"
	result = s.ssh.RunCommand(chmodCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to make Docker Compose executable")
		return false
	}

	// Create Docker directories
	mkdirCmd := "mkdir -p /opt/docker/{applications,volumes,networks}"
	result = s.ssh.RunCommand(mkdirCmd, false)
	if !result.Success {
		fmt.Println("⚠️  Could not create Docker directories")
	}

	// Test Docker installation
	testCmd := "docker --version && docker-compose --version"
	result = s.ssh.RunCommand(testCmd, false)
	if result.Success {
		fmt.Println("✅ Docker and Docker Compose installed successfully!")
		fmt.Println(result.Stdout)
		return true
	}

	fmt.Println("❌ Docker installation test failed")
	return false
}

func (s *Service) Deploy(composeFile string) bool {
	fmt.Printf("🚀 Deploying Docker Compose: %s\n", composeFile)

	// Create example docker-compose.yml
	composeContent := `version: '3.8'
services:
  whoami:
    image: traefik/whoami
    container_name: whoami
    ports:
      - "8080:80"
    restart: unless-stopped`

	// Write docker-compose.yml
	if !s.ssh.WriteFile(composeContent, "/opt/docker/docker-compose.yml") {
		fmt.Println("❌ Failed to create docker-compose.yml")
		return false
	}

	// Deploy
	deployCmd := "cd /opt/docker && docker-compose up -d"
	result := s.ssh.RunCommand(deployCmd, false)
	if result.Success {
		fmt.Println("✅ Docker Compose deployed successfully!")
		return true
	}

	fmt.Println("❌ Docker Compose deployment failed")
	return false
}