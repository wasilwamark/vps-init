.PHONY: install clean test

# Install to system (build and install globally)
install:
	@echo "Building vps-init..."
	go build -o /tmp/vps-init ./cmd/vps-init
	@echo "Installing to /usr/local/bin..."
	sudo cp /tmp/vps-init /usr/local/bin/vps-init && sudo chmod +x /usr/local/bin/vps-init
	@rm -f /tmp/vps-init
	@echo "✅ Installation complete! Run: vps-init --help"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o vps-init-linux-amd64 ./cmd/vps-init
	GOOS=linux GOARCH=arm64 go build -o vps-init-linux-arm64 ./cmd/vps-init
	GOOS=darwin GOARCH=amd64 go build -o vps-init-darwin-amd64 ./cmd/vps-init
	GOOS=darwin GOARCH=arm64 go build -o vps-init-darwin-arm64 ./cmd/vps-init
	GOOS=windows GOARCH=amd64 go build -o vps-init-windows-amd64.exe ./cmd/vps-init
	@echo "All builds completed"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ /tmp/vps-init 2>/dev/null || true

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

