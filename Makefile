.PHONY: build install clean test

# Build the CLI (clean build)
build:
	@echo "Building vps-init..."
	@mkdir -p bin
	go build -o bin/vps-init ./cmd/vps-init
	@echo "Build complete: bin/vps-init"

# Install to system (clean build + install globally only)
install: clean build
	@echo "🗑️  Removing any existing local installations..."
	@rm -f ~/bin/vps-init ~/.local/bin/vps-init 2>/dev/null || true
	@echo "📦 Installing vps-init to /usr/local/bin (global only)..."
	@if sudo cp bin/vps-init /usr/local/bin/ && sudo chmod +x /usr/local/bin/vps-init; then \
		echo "✅ Global installation successful"; \
		echo ""; \
		echo "Installation complete!"; \
		echo ""; \
		echo "Installed to:"; \
		echo "  - /usr/local/bin/vps-init (global) ✓"; \
		echo ""; \
		echo "You can now run: vps-init --help"; \
	else \
		echo "❌ Global installation failed (needs sudo privileges)"; \
		exit 1; \
	fi

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o bin/vps-init-linux-amd64 ./cmd/vps-init
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -o bin/vps-init-linux-arm64 ./cmd/vps-init
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -o bin/vps-init-darwin-amd64 ./cmd/vps-init
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o bin/vps-init-darwin-arm64 ./cmd/vps-init
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o bin/vps-init-windows-amd64.exe ./cmd/vps-init
	@echo "All builds completed in bin/"

# Quick install (build and copy to global path first, then local if needed)
install-local: build
	@echo "🗑️  Removing any existing local installations to avoid conflicts..."
	@rm -f ~/bin/vps-init ~/.local/bin/vps-init 2>/dev/null || true
	@echo "📦 Installing to global location first..."
	@if cp bin/vps-init /usr/local/bin/ && chmod +x /usr/local/bin/vps-init 2>/dev/null; then \
		echo "✅ Global installation successful"; \
	elif cp bin/vps-init ~/bin/ && chmod +x ~/bin/vps-init 2>/dev/null; then \
		echo "⚠️  Global installation failed, installed to ~/bin/vps-init"; \
	elif mkdir -p ~/.local/bin && cp bin/vps-init ~/.local/bin/ && chmod +x ~/.local/bin/vps-init 2>/dev/null; then \
		echo "⚠️  Global and ~/bin installation failed, installed to ~/.local/bin/vps-init"; \
		echo "💡 Add ~/.local/bin to your PATH: export PATH=\"$$HOME/.local/bin:$$PATH\""; \
	else \
		echo "❌ All installation locations failed"; \
		echo "💡 Add $(PWD)/bin to your PATH: export PATH=\"$(PWD)/bin:$$PATH\""; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	rm -rf bin/

# Clean local installations (remove only local copies, keep global)
clean-local:
	@echo "🗑️  Removing local installations..."
	@rm -f ~/bin/vps-init ~/.local/bin/vps-init 2>/dev/null || true
	@echo "✅ Local installations removed"
	@echo "💡 Global installation at /usr/local/bin/vps-init remains untouched"

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Development build with debug info
dev:
	go build -gcflags="all=-N -l" -o bin/vps-init-dev ./cmd/vps-init

# Install development version
install-dev: dev
	@echo "🗑️  Removing any existing local installations to avoid conflicts..."
	@rm -f ~/bin/vps-init ~/.local/bin/vps-init 2>/dev/null || true
	@echo "🔧 Installing development version to global location first..."
	@if cp bin/vps-init-dev /usr/local/bin/vps-init && chmod +x /usr/local/bin/vps-init 2>/dev/null; then \
		echo "✅ Development version installed globally"; \
	elif cp bin/vps-init-dev ~/bin/vps-init && chmod +x ~/bin/vps-init 2>/dev/null; then \
		echo "⚠️  Global installation failed, installed to ~/bin/vps-init"; \
	else \
		echo "❌ Installation failed"; \
		echo "💡 Add $(PWD)/bin to your PATH: export PATH=\"$(PWD)/bin:$$PATH\""; \
		exit 1; \
	fi