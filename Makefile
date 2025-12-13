.PHONY: build install clean test

# Build the CLI
build:
	go build -o bin/vps-init ./cmd/vps-init

# Install to system
install: build
	sudo cp bin/vps-init /usr/local/bin/
	sudo chmod +x /usr/local/bin/vps-init

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

# Quick install (build and copy to local path)
install-local: build
	cp bin/vps-init /usr/local/bin/ 2>/dev/null || cp bin/vps-init ~/bin/ 2>/dev/null || cp bin/vps-init ~/.local/bin/ 2>/dev/null || echo "Add $(PWD)/bin to your PATH"

# Clean build artifacts
clean:
	rm -rf bin/

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
	cp bin/vps-init-dev /usr/local/bin/vps-init 2>/dev/null || cp bin/vps-init-dev ~/bin/vps-init 2>/dev/null || echo "Add bin to PATH"