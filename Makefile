# CodeHunter v2.5 - Build System
# Made with ❤️ by Albert.C @yz9yt
# https://github.com/Acorzo1983/Codehunter
# 🏴‍☠️ Exclusive for Kali Linux & Linux Distributions

.PHONY: check-os build install uninstall clean test help

# Check if running on supported OS
check-os:
	@echo "🔍 Checking operating system..."
	@case "$$(uname -s)" in \
		Linux*) echo "✅ Linux detected - Perfect for Bug Bounty! 🐧" ;; \
		Darwin*) echo "✅ macOS detected - Compatible! 🍎" ;; \
		CYGWIN*|MINGW*|MSYS*) echo "❌ Windows detected - Not supported! Use WSL2 instead" && exit 1 ;; \
		*) echo "❌ Unknown OS - Linux/macOS only" && exit 1 ;; \
	esac

# Build CodeHunter binary
build: check-os
	@echo "🔨 Building CodeHunter..."
	@go build -ldflags="-s -w" -o codehunter main.go
	@echo "✅ Build complete: ./codehunter"

# Install system-wide
install: build
	@echo "📦 Installing CodeHunter system-wide..."
	@sudo cp codehunter /usr/local/bin/
	@sudo mkdir -p /usr/share/codehunter/patterns
	@sudo cp patterns/* /usr/share/codehunter/patterns/
	@sudo chmod +x /usr/local/bin/codehunter
	@echo "✅ CodeHunter installed!"
	@echo "🎯 Run: codehunter -r secrets.txt -l urls.txt"

# Uninstall
uninstall:
	@echo "🗑️ Uninstalling CodeHunter..."
	@sudo rm -f /usr/local/bin/codehunter
	@sudo rm -rf /usr/share/codehunter
	@echo "✅ CodeHunter uninstalled"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build files..."
	@rm -f codehunter
	@echo "✅ Clean complete"

# Test with example URLs
test: build
	@echo "🧪 Testing CodeHunter..."
	@./codehunter -r patterns/secrets.txt -l examples/urls.txt -v
	@echo "✅ Test complete"

# Quick development build
dev: 
	@go build -o codehunter main.go
	@echo "🚀 Dev build ready: ./codehunter"

# Show help
help:
	@echo "🏴‍☠️ CodeHunter v2.5 - Ultra-Fast Bug Bounty Scanner"
	@echo ""
	@echo "📋 Available commands:"
	@echo "  make build     - Build CodeHunter binary"
	@echo "  make install   - Install system-wide"
	@echo "  make uninstall - Remove from system" 
	@echo "  make test      - Test with examples"
	@echo "  make clean     - Clean build files"
	@echo "  make dev       - Quick dev build"
	@echo "  make help      - Show this help"
	@echo ""
	@echo "🎯 Usage examples:"
	@echo "  codehunter -r secrets.txt -l urls.txt -o found.txt"
	@echo "  katana -u tesla.com | codehunter -r api_endpoints.txt"
	@echo "  proxychains codehunter -r admin_panels.txt -l urls.txt"
	@echo ""
	@echo "Made with ❤️ by Albert.C @yz9yt"

# Default target
all: build
