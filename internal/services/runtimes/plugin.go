package runtimes

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

type RuntimeInfo struct {
	Name        string
	Command     string
	VersionFlag string
	InstallCmd  string
	Versions    []string
	DefaultVer  string
}

type LanguageRuntime struct {
	Language  string
	Runtimes  map[string]*RuntimeInfo
	Installer map[string]string
}

func (p *Plugin) Name() string {
	return "runtime"
}

func (p *Plugin) Description() string {
	return "Manage programming language runtime (Node.js, Python, Go, Java, etc.)"
}

func (p *Plugin) Author() string {
	return "VPS-Init"
}

func (p *Plugin) Version() string {
	return "0.0.1"
}

func (p *Plugin) Initialize(config map[string]interface{}) error {
	return nil
}

func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

func (p *Plugin) Dependencies() []string {
	return []string{"curl", "wget", "git"}
}

func (p *Plugin) GetRootCommand() *cobra.Command {
	return nil
}

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install a language runtime",
			Handler:     p.installHandler,
		},
		{
			Name:        "list",
			Description: "List available and installed runtime",
			Handler:     p.listHandler,
		},
		{
			Name:        "use",
			Description: "Switch to a specific version of a runtime",
			Handler:     p.useHandler,
		},
		{
			Name:        "remove",
			Description: "Remove a runtime version",
			Handler:     p.removeHandler,
		},
		{
			Name:        "status",
			Description: "Show current active runtime",
			Handler:     p.statusHandler,
		},
		{
			Name:        "update",
			Description: "Update runtime version managers",
			Handler:     p.updateHandler,
		},
	}
}

func (p *Plugin) installHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: runtime install <language> <version> [options]\n\nExample: runtime install node 18\nExample: runtime install python 3.11\nExample: runtime install go 1.21")
	}

	language := strings.ToLower(args[0])
	version := args[1]

	pass := getSudoPass(flags)

	switch language {
	case "node", "nodejs", "node.js":
		return p.installNode(ctx, conn, version, pass)
	case "python", "py", "python3":
		return p.installPython(ctx, conn, version, pass)
	case "go", "golang":
		return p.installGo(ctx, conn, version, pass)
	case "java", "jdk":
		return p.installJava(ctx, conn, version, pass)
	case "rust":
		return p.installRust(ctx, conn, version, pass)
	case "php":
		return p.installPHP(ctx, conn, version, pass)
	case "ruby":
		return p.installRuby(ctx, conn, version, pass)
	case "dotnet", ".net":
		return p.installDotNet(ctx, conn, version, pass)
	default:
		return fmt.Errorf("unsupported language: %s. Supported languages: node, python, go, java, rust, php, ruby, dotnet", language)
	}
}

func (p *Plugin) installNode(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Node.js %s...\n", version)

	// Check if nvm exists
	if res := conn.RunCommand("command -v nvm", false); !res.Success {
		fmt.Println("🔧 Installing NVM (Node Version Manager)...")
		installCmd := `curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash`
		if res := conn.RunCommand(installCmd, false); !res.Success {
			return fmt.Errorf("failed to install NVM: %s", res.Stderr)
		}

		// Add NVM to shell profile
		profileCmd := `echo 'export NVM_DIR="$HOME/.nvm"' >> ~/.bashrc && echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"' >> ~/.bashrc && echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"' >> ~/.bashrc`
		if res := conn.RunCommand(profileCmd, false); !res.Success {
			fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
		}
	}

	// Install Node.js using nvm
	installCmd := fmt.Sprintf(`bash -c 'source ~/.nvm/nvm.sh && nvm install %s && nvm use %s && nvm alias default %s'`, version, version, version)
	if res := conn.RunCommand(installCmd, false); !res.Success {
		return fmt.Errorf("failed to install Node.js %s: %s", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf(`bash -c 'source ~/.nvm/nvm.sh && node --version && npm --version'`)
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Node.js installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Node.js %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installPython(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Python %s...\n", version)

	// Install uv if not exists
	if res := conn.RunCommand("command -v uv", false); !res.Success {
		fmt.Println("🔧 Installing uv...")
		installDeps := `apt-get update 2>/dev/null || true && apt-get install -y curl`
		if res := conn.RunSudo(installDeps, pass); !res.Success {
			return fmt.Errorf("failed to install dependencies for uv: %s", res.Stderr)
		}

		installCmd := `curl -LsSf https://astral.sh/uv/install.sh | sh`
		if res := conn.RunCommand(installCmd, false); !res.Success {
			return fmt.Errorf("failed to install uv: %s", res.Stderr)
		}

		// Add uv to PATH (uv installs to ~/.local/bin)
		profileCmd := `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc`
		if res := conn.RunCommand(profileCmd, false); !res.Success {
			fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
		}

		// Export PATH for current session
		exportCmd := `export PATH="$HOME/.local/bin:$PATH"`
		conn.RunCommand(exportCmd, false)
	}

	// Install Python version using uv
	installCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv python install %s'`, version)
	if res := conn.RunCommand(installCmd, false); !res.Success {
		return fmt.Errorf("failed to install Python %s: %s", version, res.Stderr)
	}

	// Set the Python version as default
	pinCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv python pin %s'`, version)
	if res := conn.RunCommand(pinCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to pin Python %s: %s\n", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv run python --version && uv run pip --version'`)
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Python installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Python %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installGo(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Go %s...\n", version)

	// Format version string for Go download URLs
	goVersion := version
	if !strings.Contains(version, ".") {
		goVersion = version + ".22.0"  // Default to latest patch for major version
	} else if len(strings.Split(version, ".")) == 2 {
		goVersion = version + ".0"  // Add patch version if missing
	}

	// Download and install Go
	downloadCmd := fmt.Sprintf(`wget https://go.dev/dl/go%s.linux-amd64.tar.gz -O /tmp/go%s.linux-amd64.tar.gz`, goVersion, goVersion)
	if res := conn.RunCommand(downloadCmd, false); !res.Success {
		return fmt.Errorf("failed to download Go: %s", res.Stderr)
	}

	// Extract Go to /usr/local
	extractCmd := fmt.Sprintf(`sudo tar -C /usr/local -xzf /tmp/go%s.linux-amd64.tar.gz`, goVersion)
	if res := conn.RunSudo(extractCmd, pass); !res.Success {
		return fmt.Errorf("failed to extract Go: %s", res.Stderr)
	}

	// Remove the tar file
	cleanupCmd := fmt.Sprintf(`rm /tmp/go%s.linux-amd64.tar.gz`, goVersion)
	conn.RunCommand(cleanupCmd, false)

	// Add Go to PATH
	profileCmd := `echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && echo 'export GOPATH=$HOME/go' >> ~/.bashrc && echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc`
	if res := conn.RunCommand(profileCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
	}

	// Verify installation
	verifyCmd := `/usr/local/go/bin/go version`
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Go installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Go %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
		fmt.Println("📝 Note: Run 'source ~/.bashrc' or logout/login to update PATH for go command")
	}

	return nil
}

func (p *Plugin) installJava(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Java %s...\n", version)

	// Install OpenJDK
	var installCmd string
	if strings.HasPrefix(version, "8") {
		installCmd = "apt-get update 2>/dev/null || true && apt-get install -y openjdk-8-jdk"
	} else if strings.HasPrefix(version, "11") {
		installCmd = "apt-get update 2>/dev/null || true && apt-get install -y openjdk-11-jdk"
	} else if strings.HasPrefix(version, "17") {
		installCmd = "apt-get update 2>/dev/null || true && apt-get install -y openjdk-17-jdk"
	} else if strings.HasPrefix(version, "21") {
		installCmd = "apt-get update 2>/dev/null || true && apt-get install -y openjdk-21-jdk"
	} else {
		return fmt.Errorf("unsupported Java version: %s. Supported versions: 8, 11, 17, 21", version)
	}

	// Install Java with single command (like other runtimes)
	if res := conn.RunSudo(installCmd, pass); !res.Success {
		return fmt.Errorf("failed to install Java %s: %s", version, res.Stderr)
	}

	// Set JAVA_HOME
	homeCmd := `echo 'export JAVA_HOME=/usr/lib/jvm/java-'$(ls /usr/lib/jvm | grep openjdk | head -n 1 | cut -d'-' -f2)'-openjdk-amd64' >> ~/.bashrc`
	if res := conn.RunCommand(homeCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to set JAVA_HOME: %s\n", res.Stderr)
	}

	// Verify installation
	verifyCmd := `bash -c 'source ~/.bashrc && java -version && javac -version'`
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Java installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Java %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installRust(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Rust %s...\n", version)

	// Install Rust using rustup
	installCmd := `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y`
	if res := conn.RunCommand(installCmd, false); !res.Success {
		return fmt.Errorf("failed to install Rust: %s", res.Stderr)
	}

	// Add cargo to PATH
	profileCmd := `echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc`
	if res := conn.RunCommand(profileCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
	}

	// Export PATH for current session
	exportCmd := `export PATH="$HOME/.cargo/bin:$PATH"`
	conn.RunCommand(exportCmd, false)

	// Verify installation
	verifyCmd := `bash -c 'export PATH="$HOME/.cargo/bin:$PATH" && rustc --version && cargo --version'`
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Rust installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Rust installed successfully!\n")
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installPHP(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing PHP %s...\n", version)

	// Add PPA for newer PHP versions, but handle gracefully if it fails
	if version > "8.0" {
		ppaCmd := `apt-get install -y software-properties-common && add-apt-repository -y ppa:ondrej/php 2>/dev/null || true`
		if res := conn.RunSudo(ppaCmd, pass); !res.Success {
			fmt.Printf("⚠️  Failed to add PHP PPA, trying with default repositories\n")
		}
	}

	// Install PHP
	var installCmd string
	if version == "8" {
		installCmd = "apt-get update 2>/dev/null || true && apt-get install -y php8.1 php8.1-cli php8.1-fpm php8.1-mbstring php8.1-xml php8.1-curl"
	} else if strings.HasPrefix(version, "8.") {
		installCmd = fmt.Sprintf("apt-get update 2>/dev/null || true && apt-get install -y php%s php%s-cli php%s-fpm php%s-mbstring php%s-xml php%s-curl", version, version, version, version, version, version)
	} else {
		installCmd = fmt.Sprintf("apt-get update 2>/dev/null || true && apt-get install -y php%s php%s-cli php%s-fpm php%s-mbstring php%s-xml php%s-curl", version, version, version, version, version, version)
	}
	if res := conn.RunSudo(installCmd, pass); !res.Success {
		// Fallback to default PHP version if specific version fails
		fmt.Printf("⚠️  PHP %s not available, trying with default PHP version...\n", version)
		fallbackCmd := "apt-get update 2>/dev/null || true && apt-get install -y php php-cli php-fpm php-mbstring php-xml php-curl"
		if res := conn.RunSudo(fallbackCmd, pass); !res.Success {
			return fmt.Errorf("failed to install PHP: %s", res.Stderr)
		}
	}

	// Verify installation
	verifyCmd := fmt.Sprintf("php%s --version", version)
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		// Fallback to generic php command if version-specific fails
		verifyCmd = "php --version"
		if res := conn.RunCommand(verifyCmd, false); !res.Success {
			fmt.Printf("⚠️  Failed to verify PHP installation: %s\n", res.Stderr)
		} else {
			fmt.Printf("✅ PHP installed successfully!\n")
			fmt.Println(res.Stdout)
		}
	} else {
		fmt.Printf("✅ PHP %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installRuby(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Ruby %s...\n", version)

	// Format Ruby version for rbenv
	rubyVersion := version
	if version == "3" {
		rubyVersion = "3.3.0"  // Default to latest stable 3.x
	} else if len(strings.Split(version, ".")) == 1 {
		rubyVersion = version + ".0.0"  // Add minor and patch if missing
	} else if len(strings.Split(version, ".")) == 2 {
		rubyVersion = version + ".0"  // Add patch version if missing
	}

	// Install Ruby using rbenv
	installDeps := `apt-get update 2>/dev/null || true && apt-get install -y autoconf bison build-essential libssl-dev libyaml-dev libreadline6-dev zlib1g-dev libncurses5-dev libffi-dev libgdbm-dev git`
	if res := conn.RunSudo(installDeps, pass); !res.Success {
		return fmt.Errorf("failed to install Ruby build dependencies: %s", res.Stderr)
	}

	// Install rbenv (check if already installed first)
	if res := conn.RunCommand("test -d ~/.rbenv", false); !res.Success {
		rbenvCmd := `git clone https://github.com/rbenv/rbenv.git ~/.rbenv && git clone https://github.com/rbenv/ruby-build.git ~/.rbenv/plugins/ruby-build`
		if res := conn.RunCommand(rbenvCmd, false); !res.Success {
			return fmt.Errorf("failed to install rbenv: %s", res.Stderr)
		}
	}

	// Add rbenv to PATH
	profileCmd := `echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.bashrc && echo 'eval "$(rbenv init -)"' >> ~/.bashrc`
	if res := conn.RunCommand(profileCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
	}

	// Install Ruby version
	installCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && rbenv install %s --skip-existing && rbenv global %s'`, rubyVersion, rubyVersion)
	if res := conn.RunCommand(installCmd, false); !res.Success {
		return fmt.Errorf("failed to install Ruby %s: %s", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && ruby --version && gem --version'`)
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Ruby installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Ruby %s installed successfully!\n", rubyVersion)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installDotNet(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing .NET %s...\n", version)

	// Detect Ubuntu version
	ubuntuVersion := "20.04"
	if res := conn.RunCommand(`lsb_release -rs | cut -d. -f1`, false); res.Success {
		ver := strings.TrimSpace(res.Stdout)
		if ver == "22" || ver == "24" {
			ubuntuVersion = ver + ".04"
		}
	}

	// Add Microsoft package repository
	repoCmd := fmt.Sprintf(`apt-get update 2>/dev/null || true && apt-get install -y wget && wget https://packages.microsoft.com/config/ubuntu/%s/packages-microsoft-prod.deb -O packages-microsoft-prod.deb && dpkg -i packages-microsoft-prod.deb`, ubuntuVersion)
	if res := conn.RunSudo(repoCmd, pass); !res.Success {
		return fmt.Errorf("failed to add Microsoft repository: %s", res.Stderr)
	}

	// Install .NET SDK
	dotnetVersion := version
	if len(strings.Split(version, ".")) == 1 {
		dotnetVersion = version + ".0"
	}
	installCmd := fmt.Sprintf("apt-get update 2>/dev/null || true && apt-get install -y dotnet-sdk-%s", dotnetVersion)
	if res := conn.RunSudo(installCmd, pass); !res.Success {
		return fmt.Errorf("failed to install .NET %s: %s", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf("dotnet --version")
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify .NET installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ .NET %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) listHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("📋 Available Runtimes:")
	fmt.Println()

	runtimes := []struct {
		name   string
		cmd    string
		extra  string
	}{
		{"Node.js", "bash -c 'source ~/.nvm/nvm.sh && nvm list'", ""},
		{"Python", "bash -c 'export PATH=\"$HOME/.local/bin:$PATH\" && uv python list'", ""},
		{"Go", "go version", ""},
		{"Java", "java -version 2>&1 && javac -version 2>&1", ""},
		{"Rust", "rustc --version", ""},
		{"PHP", "php --version", ""},
		{"Ruby", "bash -c 'export PATH=\"$HOME/.rbenv/bin:$PATH\" && eval \"$(rbenv init -)\" && ruby --version'", ""},
		{".NET", "dotnet --version", ""},
	}

	for _, rt := range runtimes {
		fmt.Printf("=== %s ===\n", rt.name)
		if res := conn.RunCommand(rt.cmd, false); res.Success {
			if rt.name == "Node.js" && res.Stdout == "" {
				fmt.Println("  NVM not installed or no Node.js versions found")
			} else if rt.name == "Python" && res.Stdout == "" {
				fmt.Println("  uv not installed or no Python versions found")
			} else if rt.name == "Ruby" && res.Stdout == "" {
				fmt.Println("  rbenv not installed or no Ruby versions found")
			} else {
				output := strings.TrimSpace(res.Stdout)
				if output != "" {
					for _, line := range strings.Split(output, "\n") {
						fmt.Printf("  %s\n", line)
					}
				}
			}
		} else {
			fmt.Printf("  Not installed or not in PATH\n")
		}
		if rt.extra != "" {
			if res := conn.RunCommand(rt.extra, false); res.Success && strings.TrimSpace(res.Stdout) != "" {
				for _, line := range strings.Split(strings.TrimSpace(res.Stdout), "\n") {
					fmt.Printf("  %s\n", line)
				}
			}
		}
		fmt.Println()
	}

	return nil
}

func (p *Plugin) useHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: runtime use <language> <version>")
	}

	language := strings.ToLower(args[0])
	version := args[1]

	switch language {
	case "node", "nodejs", "node.js":
		return p.useNode(ctx, conn, version)
	case "python", "py", "python3":
		return p.usePython(ctx, conn, version)
	case "go", "golang":
		fmt.Println("Go version switching is not supported. Please reinstall the desired version.")
		return nil
	case "java", "jdk":
		fmt.Println("Java version switching requires update-alternatives. Please use: sudo update-alternatives --config java")
		return nil
	default:
		return fmt.Errorf("runtime switching not supported for %s", language)
	}
}

func (p *Plugin) useNode(ctx context.Context, conn *ssh.Connection, version string) error {
	fmt.Printf("🔄 Switching to Node.js %s...\n", version)

	cmd := fmt.Sprintf(`bash -c 'source ~/.nvm/nvm.sh && nvm use %s && nvm alias default %s'`, version, version)
	if res := conn.RunCommand(cmd, false); !res.Success {
		return fmt.Errorf("failed to switch Node.js version: %s", res.Stderr)
	}

	verifyCmd := fmt.Sprintf(`bash -c 'source ~/.nvm/nvm.sh && node --version'`)
	res := conn.RunCommand(verifyCmd, false)
	if !res.Success {
		return fmt.Errorf("failed to verify Node.js version: %s", res.Stderr)
	}

	fmt.Printf("✅ Now using Node.js %s\n", strings.TrimSpace(res.Stdout))
	return nil
}

func (p *Plugin) usePython(ctx context.Context, conn *ssh.Connection, version string) error {
	fmt.Printf("🔄 Switching to Python %s...\n", version)

	cmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv python pin %s'`, version)
	if res := conn.RunCommand(cmd, false); !res.Success {
		return fmt.Errorf("failed to switch Python version: %s", res.Stderr)
	}

	verifyCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv run python --version'`)
	res := conn.RunCommand(verifyCmd, false)
	if !res.Success {
		return fmt.Errorf("failed to verify Python version: %s", res.Stderr)
	}

	fmt.Printf("✅ Now using Python %s\n", strings.TrimSpace(res.Stdout))
	return nil
}

func (p *Plugin) removeHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: runtime remove <language> <version>")
	}

	language := strings.ToLower(args[0])
	version := args[1]
	pass := getSudoPass(flags)

	switch language {
	case "node", "nodejs", "node.js":
		cmd := fmt.Sprintf(`bash -c 'source ~/.nvm/nvm.sh && nvm uninstall %s'`, version)
		if res := conn.RunCommand(cmd, false); !res.Success {
			return fmt.Errorf("failed to uninstall Node.js %s: %s", version, res.Stderr)
		}
		fmt.Printf("✅ Node.js %s uninstalled\n", version)

	case "python", "py", "python3":
		cmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv python uninstall %s'`, version)
		if res := conn.RunCommand(cmd, false); !res.Success {
			return fmt.Errorf("failed to uninstall Python %s: %s", version, res.Stderr)
		}
		fmt.Printf("✅ Python %s uninstalled\n", version)

	case "go", "golang":
		// Check if Go is installed in /usr/local/go
		checkCmd := fmt.Sprintf(`/usr/local/go/bin/go version | grep 'go%s'`, version)
		if res := conn.RunCommand(checkCmd, false); !res.Success {
			return fmt.Errorf("Go %s not found in /usr/local/go", version)
		}

		// Remove Go installation
		removeCmd := fmt.Sprintf(`sudo rm -rf /usr/local/go%s && sudo rm -rf /usr/local/go`, version)
		if res := conn.RunSudo(removeCmd, pass); !res.Success {
			return fmt.Errorf("failed to remove Go %s: %s", version, res.Stderr)
		}

		// Remove PATH entries from bashrc
		cleanupCmd := `sed -i '/export PATH=\$PATH:\/usr\/local\/go\/bin/d' ~/.bashrc && sed -i '/export GOPATH=\$HOME\/go/d' ~/.bashrc && sed -i '/export PATH=\$PATH:\$GOPATH\/bin/d' ~/.bashrc`
		conn.RunCommand(cleanupCmd, false)

		fmt.Printf("✅ Go %s uninstalled\n", version)

	case "java", "jdk":
		// Find Java installation
		findCmd := fmt.Sprintf(`ls /usr/lib/jvm/ | grep -E 'java-%s-openjdk' | head -n 1`, version)
		res := conn.RunCommand(findCmd, false)
		if !res.Success || strings.TrimSpace(res.Stdout) == "" {
			return fmt.Errorf("Java %s not found in /usr/lib/jvm", version)
		}

		javaDir := strings.TrimSpace(res.Stdout)

		// Remove Java installation
		removeCmd := fmt.Sprintf(`sudo rm -rf /usr/lib/jvm/%s`, javaDir)
		if res := conn.RunSudo(removeCmd, pass); !res.Success {
			return fmt.Errorf("failed to remove Java %s: %s", version, res.Stderr)
		}

		// Remove JAVA_HOME from bashrc
		cleanupCmd := `sed -i '/export JAVA_HOME/d' ~/.bashrc`
		conn.RunCommand(cleanupCmd, false)

		fmt.Printf("✅ Java %s uninstalled\n", version)

	case "rust":
		// Use rustup to uninstall
		cmd := fmt.Sprintf(`bash -c 'source ~/.bashrc && rustup self uninstall -y'`)
		if res := conn.RunCommand(cmd, false); !res.Success {
			// Fallback: manually remove rust
			removeCmd := `rm -rf ~/.cargo && rm -rf ~/.rustup`
			if res := conn.RunCommand(removeCmd, false); !res.Success {
				return fmt.Errorf("failed to remove Rust: %s", res.Stderr)
			}
		}

		// Remove PATH entries from bashrc
		cleanupCmd := `sed -i '/export PATH="\$HOME\/\.cargo\/bin:\$PATH"/d' ~/.bashrc`
		conn.RunCommand(cleanupCmd, false)

		fmt.Printf("✅ Rust uninstalled\n")

	case "php":
		// Remove PHP packages
		removeCmd := fmt.Sprintf(`sudo apt-get remove -y php%s php%s-cli php%s-fpm php%s-mbstring php%s-xml php%s-curl`, version, version, version, version, version, version)
		if res := conn.RunSudo(removeCmd, pass); !res.Success {
			return fmt.Errorf("failed to remove PHP %s: %s", version, res.Stderr)
		}

		fmt.Printf("✅ PHP %s uninstalled\n", version)

	case "ruby":
		// Use rbenv to uninstall
		cmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && rbenv uninstall %s'`, version)
		if res := conn.RunCommand(cmd, false); !res.Success {
			return fmt.Errorf("failed to uninstall Ruby %s: %s", version, res.Stderr)
		}

		fmt.Printf("✅ Ruby %s uninstalled\n", version)

	case "dotnet", ".net", "net":
		// Remove .NET SDK
		removeCmd := fmt.Sprintf(`sudo apt-get remove -y dotnet-sdk-%s`, version)
		if res := conn.RunSudo(removeCmd, pass); !res.Success {
			return fmt.Errorf("failed to remove .NET %s: %s", version, res.Stderr)
		}

		fmt.Printf("✅ .NET %s uninstalled\n", version)

	default:
		return fmt.Errorf("runtime removal not supported for %s. Supported languages: node, python, go, java, rust, php, ruby, dotnet", language)
	}

	return nil
}

func (p *Plugin) statusHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔍 Current Runtime Status:")
	fmt.Println()

	commands := []struct {
		name string
		cmd  string
	}{
		{"Node.js", `bash -c 'source ~/.nvm/nvm.sh && echo "Node: $(node --version 2>/dev/null || echo "Not found")" && echo "NPM: $(npm --version 2>/dev/null || echo "Not found")"'`},
		{"Python", `bash -c 'export PATH="$HOME/.local/bin:$PATH" && echo "Python: $(uv run python --version 2>/dev/null || echo "Not found")" && echo "Pip: $(uv run pip --version 2>/dev/null || echo "Not found")" && echo "UV: $(uv --version 2>/dev/null || echo "Not found")"'`},
		{"Go", "bash -c 'echo \"Go: $(/usr/local/go/bin/go version 2>/dev/null || /usr/bin/go version 2>/dev/null || echo \"Not found\")\"'"},
		{"Java", "bash -c 'echo \"Java: $(java -version 2>&1 | head -n 1 || echo \"Not found\")\"'"},
		{"Rust", "bash -c 'echo \"Rust: $(/home/ubuntu/.cargo/bin/rustc --version 2>/dev/null || /usr/bin/rustc --version 2>/dev/null || echo \"Not found\")\"'"},
		{"PHP", "bash -c 'echo \"PHP: $(php --version 2>/dev/null | head -n 1 || echo \"Not found\")\"'"},
		{"Ruby", `bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && echo "Ruby: $(ruby --version 2>/dev/null || echo "Not found")"'`},
		{".NET", "bash -c 'echo \".NET: $(dotnet --version 2>/dev/null || echo \"Not found\")\"'"},
	}

	for _, c := range commands {
		fmt.Printf("=== %s ===\n", c.name)
		if res := conn.RunCommand(c.cmd, false); res.Success && strings.TrimSpace(res.Stdout) != "" {
			for _, line := range strings.Split(strings.TrimSpace(res.Stdout), "\n") {
				fmt.Printf("  %s\n", line)
			}
		} else {
			fmt.Println("  Not installed or not in PATH")
		}
		fmt.Println()
	}

	return nil
}

func (p *Plugin) updateHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔄 Updating runtime version managers...")

	// Update nvm
	fmt.Println("\n📦 Updating NVM...")
	if res := conn.RunCommand(`bash -c 'source ~/.nvm/nvm.sh && nvm update-version'`, false); !res.Success {
		fmt.Println("  NVM update failed or not installed")
	}

	// Update uv
	fmt.Println("\n📦 Updating uv...")
	if res := conn.RunCommand(`bash -c 'export PATH="$HOME/.local/bin:$PATH" && uv self update'`, false); !res.Success {
		fmt.Println("  uv update failed or not installed")
	}

	// Update rbenv
	fmt.Println("\n📦 Updating rbenv...")
	if res := conn.RunCommand(`bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && rbenv update'`, false); !res.Success {
		fmt.Println("  rbenv update failed or not installed")
	}

	// Update rustup
	fmt.Println("\n📦 Updating Rust...")
	if res := conn.RunCommand(`bash -c 'rustup update'`, false); !res.Success {
		fmt.Println("  Rust update failed or not installed")
	}

	fmt.Println("\n✅ Runtime managers updated!")

	return nil
}

// Helper
func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}