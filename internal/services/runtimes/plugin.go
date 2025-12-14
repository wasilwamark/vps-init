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
	return "runtimes"
}

func (p *Plugin) Description() string {
	return "Manage programming language runtimes (Node.js, Python, Go, Java, etc.)"
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
			Description: "List available and installed runtimes",
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
			Description: "Show current active runtimes",
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
		return fmt.Errorf("usage: runtimes install <language> <version> [options]\n\nExample: runtimes install node 18\nExample: runtimes install python 3.11\nExample: runtimes install go 1.21")
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

	// Install pyenv if not exists
	if res := conn.RunCommand("command -v pyenv", false); !res.Success {
		fmt.Println("🔧 Installing pyenv...")
		installDeps := `apt-get update && apt-get install -y make build-essential libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm libncurses5-dev libncursesw5-dev xz-utils tk-dev libffi-dev liblzma-dev python3-openssl git`
		if res := conn.RunSudo(installDeps, pass); !res.Success {
			return fmt.Errorf("failed to install Python build dependencies: %s", res.Stderr)
		}

		installCmd := `curl https://pyenv.run | bash`
		if res := conn.RunCommand(installCmd, false); !res.Success {
			return fmt.Errorf("failed to install pyenv: %s", res.Stderr)
		}

		// Add pyenv to shell profile
		profileCmd := `echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc && echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc && echo 'eval "$(pyenv init -)"' >> ~/.bashrc`
		if res := conn.RunCommand(profileCmd, false); !res.Success {
			fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
		}
	}

	// Install Python version
	installCmd := fmt.Sprintf(`bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && pyenv install %s && pyenv global %s'`, version, version)
	if res := conn.RunCommand(installCmd, false); !res.Success {
		return fmt.Errorf("failed to install Python %s: %s", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf(`bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && python --version && pip --version'`)
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

	// Download and install Go
	downloadCmd := fmt.Sprintf(`wget https://go.dev/dl/go%s.linux-amd64.tar.gz -O /tmp/go%s.linux-amd64.tar.gz`, version, version)
	if res := conn.RunCommand(downloadCmd, false); !res.Success {
		return fmt.Errorf("failed to download Go: %s", res.Stderr)
	}

	// Extract Go to /usr/local
	extractCmd := fmt.Sprintf(`sudo tar -C /usr/local -xzf /tmp/go%s.linux-amd64.tar.gz`, version)
	if res := conn.RunSudo(extractCmd, pass); !res.Success {
		return fmt.Errorf("failed to extract Go: %s", res.Stderr)
	}

	// Remove the tar file
	cleanupCmd := fmt.Sprintf(`rm /tmp/go%s.linux-amd64.tar.gz`, version)
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
		installCmd = "apt-get update && apt-get install -y openjdk-8-jdk"
	} else if strings.HasPrefix(version, "11") {
		installCmd = "apt-get update && apt-get install -y openjdk-11-jdk"
	} else if strings.HasPrefix(version, "17") {
		installCmd = "apt-get update && apt-get install -y openjdk-17-jdk"
	} else if strings.HasPrefix(version, "21") {
		installCmd = "apt-get update && apt-get install -y openjdk-21-jdk"
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

	// Verify installation
	verifyCmd := `bash -c 'source ~/.bashrc && rustc --version && cargo --version'`
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

	// Add PPA for newer PHP versions
	if version > "8.0" {
		ppaCmd := `apt-get install -y software-properties-common && add-apt-repository -y ppa:ondrej/php`
		if res := conn.RunSudo(ppaCmd, pass); !res.Success {
			fmt.Printf("⚠️  Failed to add PHP PPA: %s\n", res.Stderr)
		}
	}

	// Install PHP
	installCmd := fmt.Sprintf("apt-get update && apt-get install -y php%s php%s-cli php%s-fpm php%s-mbstring php%s-xml php%s-curl", version, version, version, version, version, version)
	if res := conn.RunSudo(installCmd, pass); !res.Success {
		return fmt.Errorf("failed to install PHP %s: %s", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf("php%s --version", version)
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify PHP installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ PHP %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installRuby(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing Ruby %s...\n", version)

	// Install Ruby using rbenv
	installDeps := `apt-get update && apt-get install -y autoconf bison build-essential libssl-dev libyaml-dev libreadline6-dev zlib1g-dev libncurses5-dev libffi-dev libgdbm-dev git`
	if res := conn.RunSudo(installDeps, pass); !res.Success {
		return fmt.Errorf("failed to install Ruby build dependencies: %s", res.Stderr)
	}

	// Install rbenv
	rbenvCmd := `git clone https://github.com/rbenv/rbenv.git ~/.rbenv && git clone https://github.com/rbenv/ruby-build.git ~/.rbenv/plugins/ruby-build`
	if res := conn.RunCommand(rbenvCmd, false); !res.Success {
		return fmt.Errorf("failed to install rbenv: %s", res.Stderr)
	}

	// Add rbenv to PATH
	profileCmd := `echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.bashrc && echo 'eval "$(rbenv init -)"' >> ~/.bashrc`
	if res := conn.RunCommand(profileCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to update bashrc: %s\n", res.Stderr)
	}

	// Install Ruby version
	installCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && rbenv install %s && rbenv global %s'`, version, version)
	if res := conn.RunCommand(installCmd, false); !res.Success {
		return fmt.Errorf("failed to install Ruby %s: %s", version, res.Stderr)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf(`bash -c 'export PATH="$HOME/.rbenv/bin:$PATH" && eval "$(rbenv init -)" && ruby --version && gem --version'`)
	if res := conn.RunCommand(verifyCmd, false); !res.Success {
		fmt.Printf("⚠️  Failed to verify Ruby installation: %s\n", res.Stderr)
	} else {
		fmt.Printf("✅ Ruby %s installed successfully!\n", version)
		fmt.Println(res.Stdout)
	}

	return nil
}

func (p *Plugin) installDotNet(ctx context.Context, conn *ssh.Connection, version string, pass string) error {
	fmt.Printf("📦 Installing .NET %s...\n", version)

	// Add Microsoft package repository
	repoCmd := `apt-get update && apt-get install -y wget && wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb && dpkg -i packages-microsoft-prod.deb`
	if res := conn.RunSudo(repoCmd, pass); !res.Success {
		return fmt.Errorf("failed to add Microsoft repository: %s", res.Stderr)
	}

	// Install .NET SDK
	installCmd := fmt.Sprintf("apt-get update && apt-get install -y dotnet-sdk-%s", version)
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
		{"Python", "bash -c 'export PYENV_ROOT=\"$HOME/.pyenv\" && export PATH=\"$PYENV_ROOT/bin:$PATH\" && eval \"$(pyenv init -)\" && pyenv versions'", ""},
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
				fmt.Println("  pyenv not installed or no Python versions found")
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
		return fmt.Errorf("usage: runtimes use <language> <version>")
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

	cmd := fmt.Sprintf(`bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && pyenv global %s'`, version)
	if res := conn.RunCommand(cmd, false); !res.Success {
		return fmt.Errorf("failed to switch Python version: %s", res.Stderr)
	}

	verifyCmd := fmt.Sprintf(`bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && python --version'`)
	res := conn.RunCommand(verifyCmd, false)
	if !res.Success {
		return fmt.Errorf("failed to verify Python version: %s", res.Stderr)
	}

	fmt.Printf("✅ Now using Python %s\n", strings.TrimSpace(res.Stdout))
	return nil
}

func (p *Plugin) removeHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: runtimes remove <language> <version>")
	}

	language := strings.ToLower(args[0])
	version := args[1]

	switch language {
	case "node", "nodejs", "node.js":
		cmd := fmt.Sprintf(`bash -c 'source ~/.nvm/nvm.sh && nvm uninstall %s'`, version)
		if res := conn.RunCommand(cmd, false); !res.Success {
			return fmt.Errorf("failed to uninstall Node.js %s: %s", version, res.Stderr)
		}
		fmt.Printf("✅ Node.js %s uninstalled\n", version)
	case "python", "py", "python3":
		cmd := fmt.Sprintf(`bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && pyenv uninstall %s'`, version)
		if res := conn.RunCommand(cmd, false); !res.Success {
			return fmt.Errorf("failed to uninstall Python %s: %s", version, res.Stderr)
		}
		fmt.Printf("✅ Python %s uninstalled\n", version)
	default:
		return fmt.Errorf("runtime removal not supported for %s", language)
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
		{"Python", `bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && echo "Python: $(python --version 2>/dev/null || echo "Not found")" && echo "Pip: $(pip --version 2>/dev/null || echo "Not found")"'`},
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

	// Update pyenv
	fmt.Println("\n📦 Updating pyenv...")
	if res := conn.RunCommand(`bash -c 'export PYENV_ROOT="$HOME/.pyenv" && export PATH="$PYENV_ROOT/bin:$PATH" && eval "$(pyenv init -)" && pyenv update'`, false); !res.Success {
		fmt.Println("  pyenv update failed or not installed")
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