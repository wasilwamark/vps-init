#!/bin/bash

# Build the firewall plugin as a shared object (.so)

echo "🔨 Building firewall plugin..."

# Get the directory of this script
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

# Set Go module name
export GO111MODULE=off

# Build as plugin
go build -buildmode=plugin -o firewall.so firewall.go

if [ $? -eq 0 ]; then
    echo "✅ Plugin built successfully: $DIR/firewall.so"
    echo ""
    echo "📦 To use this plugin:"
    echo "1. Copy firewall.so to ~/.vps-init/plugins/ or /usr/local/lib/vps-init/plugins/"
    echo "2. Add to ~/.vps-init/plugins.yaml:"
    echo "   plugins:"
    echo "     firewall:"
    echo "       enabled: true"
    echo "       path: \"~/.vps-init/plugins/firewall.so\""
    echo ""
    echo "🚀 Then use:"
    echo "   vps-init user@host firewall install"
    echo "   vps-init user@host firewall allow 80"
    echo "   vps-init user@host firewall enable"
else
    echo "❌ Failed to build plugin"
    exit 1
fi