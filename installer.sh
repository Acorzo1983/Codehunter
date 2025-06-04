#!/bin/bash

# CodeHunter v2.5 - Linux Installer Script
# Made with ❤️ by Albert.C @yz9yt
# https://github.com/Acorzo1983/Codehunter
# 🏴‍☠️ Exclusive for Kali Linux & Linux Distributions

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Banner
print_banner() {
    echo -e "${BOLD}${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║                CodeHunter v2.5 Installer                ║"
    echo "║          Ultra-Fast Bug Bounty Scanner                  ║"
    echo "║              🏴‍☠️ Kali Linux Ready                       ║"
    echo "║                                                          ║"
    echo -e "║             Made with ${RED}❤️${PURPLE} by Albert.C @yz9yt             ║"
    echo "║          github.com/Acorzo1983/Codehunter               ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

check_os() {
    echo -e "${CYAN}[INFO]${NC} Checking operating system..."
    
    case "$(uname -s)" in
        Linux*)
            echo -e "${GREEN}[SUCCESS]${NC} Linux detected - Perfect for Bug Bounty! 🐧"
            if [[ -f /etc/os-release ]]; then
                OS_ID=$(grep "^ID=" /etc/os-release | cut -d'=' -f2 | tr -d '"')
                case $OS_ID in
                    kali) echo -e "${GREEN}[INFO]${NC} Running on Kali Linux 🏴‍☠️" ;;
                    parrot) echo -e "${GREEN}[INFO]${NC} Running on Parrot OS 🦜" ;;
                    ubuntu|debian) echo -e "${GREEN}[INFO]${NC} Running on $OS_ID" ;;
                    arch) echo -e "${GREEN}[INFO]${NC} Running on Arch Linux" ;;
                    *) echo -e "${YELLOW}[INFO]${NC} Running on $OS_ID Linux" ;;
                esac
            fi
            ;;
        Darwin*)
            echo -e "${GREEN}[SUCCESS]${NC} macOS detected - Compatible! 🍎"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            echo -e "${RED}[ERROR]${NC} Windows detected - Not supported!"
            echo -e "${YELLOW}[TIP]${NC} Use WSL2 with Kali Linux instead"
            exit 1
            ;;
        *)
            echo -e "${RED}[ERROR]${NC} Unknown OS - Linux/macOS only"
            exit 1
            ;;
    esac
}

check_dependencies() {
    echo -e "\n${CYAN}[INFO]${NC} Checking dependencies..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}[ERROR]${NC} Go not found!"
        echo -e "${YELLOW}[INSTALL]${NC} Install Go:"
        echo "  Ubuntu/Debian: sudo apt install golang-go"
        echo "  Arch: sudo pacman -S go"
        echo "  macOS: brew install go"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    echo -e "${GREEN}[FOUND]${NC} Go $GO_VERSION ✓"
    
    # Check Git
    if ! command -v git &> /dev/null; then
        echo -e "${RED}[ERROR]${NC} Git not found!"
        echo -e "${YELLOW}[INSTALL]${NC} Install Git: sudo apt install git"
        exit 1
    fi
    echo -e "${GREEN}[FOUND]${NC} Git ✓"
    
    # Check Make
    if ! command -v make &> /dev/null; then
        echo -e "${YELLOW}[WARN]${NC} Make not found"
        echo -e "${CYAN}[INFO]${NC} Installing without Make..."
        USE_MAKE=false
    else
        echo -e "${GREEN}[FOUND]${NC} Make ✓"
        USE_MAKE=true
    fi
}

download_codehunter() {
    echo -e "\n${CYAN}[INFO]${NC} Downloading CodeHunter..."
    
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    if git clone https://github.com/Acorzo1983/Codehunter.git; then
        echo -e "${GREEN}[SUCCESS]${NC} CodeHunter downloaded ✓"
        cd Codehunter
    else
        echo -e "${RED}[ERROR]${NC} Failed to download CodeHunter"
        echo -e "${YELLOW}[TIP]${NC} Check internet connection and GitHub access"
        exit 1
    fi
}

build_and_install() {
    echo -e "\n${CYAN}[INFO]${NC} Building and installing CodeHunter..."
    
    if [[ "$USE_MAKE" == true ]]; then
        echo -e "${BLUE}[BUILD]${NC} Using Makefile..."
        make install
    else
        echo -e "${BLUE}[BUILD]${NC} Manual installation..."
        
        # Build
        go build -ldflags="-s -w" -o codehunter main.go
        
        # Install binary
        sudo cp codehunter /usr/local/bin/
        sudo chmod +x /usr/local/bin/codehunter
        
        # Install patterns
        sudo mkdir -p /usr/share/codehunter/patterns
        sudo cp patterns/* /usr/share/codehunter/patterns/
        
        echo -e "${GREEN}[SUCCESS]${NC} CodeHunter installed manually ✓"
    fi
}

verify_installation() {
    echo -e "\n${CYAN}[INFO]${NC} Verifying installation..."
    
    if command -v codehunter &> /dev/null; then
        echo -e "${GREEN}[SUCCESS]${NC} CodeHunter installed successfully! ✓"
        
        # Test basic functionality
        echo -e "${BLUE}[TEST]${NC} Testing basic functionality..."
        if codehunter -b=false 2>/dev/null; then
            echo -e "${GREEN}[SUCCESS]${NC} CodeHunter working correctly! ✓"
        else
            echo -e "${YELLOW}[WARN]${NC} CodeHunter installed but test failed"
        fi
    else
        echo -e "${RED}[ERROR]${NC} Installation failed"
        exit 1
    fi
}

show_usage() {
    echo -e "\n${BOLD}${GREEN}🎉 INSTALLATION COMPLETE! 🎉${NC}"
    echo -e "\n${BOLD}${YELLOW}🎯 Quick Start:${NC}"
    echo -e "${BLUE}Basic scan:${NC}"
    echo "  codehunter -r secrets.txt -l urls.txt -o found.txt"
    echo ""
    echo -e "${BLUE}Pipe with Bug Bounty tools:${NC}"
    echo "  katana -u tesla.com | codehunter -r api_endpoints.txt"
    echo "  subfinder -d tesla.com | httpx | codehunter -r secrets.txt"
    echo "  waybackurls tesla.com | codehunter -r admin_panels.txt"
    echo ""
    echo -e "${BLUE}With proxychains:${NC}"
    echo "  proxychains codehunter -r patterns.txt -l urls.txt"
    echo ""
    echo -e "${BOLD}${PURPLE}📋 Available Patterns:${NC}"
    echo "  • secrets.txt      - API keys, tokens, credentials"
    echo "  • api_endpoints.txt - REST APIs, endpoints"
    echo "  • admin_panels.txt  - Admin areas, panels"
    echo "  • js_secrets.txt   - JavaScript secrets"
    echo "  • files.txt        - Sensitive files"
    echo "  • custom.txt       - Your custom patterns"
    echo ""
    echo -e "${BOLD}${CYAN}🏴‍☠️ Made with ❤️ by Albert.C @yz9yt 🏴‍☠️${NC}"
    echo -e "${PURPLE}GitHub: https://github.com/Acorzo1983/Codehunter${NC}"
    echo -e "${PURPLE}Twitter: @yz9yt${NC}"
    echo ""
    echo -e "${BOLD}🎯 Happy Bug Hunting! 🎯${NC}"
}

cleanup() {
    echo -e "\n${CYAN}[INFO]${NC} Cleaning up temporary files..."
    cd /
    rm -rf "$TEMP_DIR" 2>/dev/null || true
}

main() {
    print_banner
    check_os
    check_dependencies
    download_codehunter
    build_and_install
    verify_installation
    cleanup
    show_usage
}

# Handle interrupts
trap cleanup EXIT

# Run main function
main "$@"
