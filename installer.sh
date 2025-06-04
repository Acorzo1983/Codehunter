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

# Variables
CURRENT_DIR=$(pwd)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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
        echo -e "${YELLOW}[WARN]${NC} Make not found - Using manual build"
        USE_MAKE=false
    else
        echo -e "${GREEN}[FOUND]${NC} Make ✓"
        USE_MAKE=true
    fi
    
    # Check sudo
    if ! command -v sudo &> /dev/null; then
        echo -e "${YELLOW}[WARN]${NC} sudo not found - Manual install required"
        USE_SUDO=false
    else
        echo -e "${GREEN}[FOUND]${NC} sudo ✓"
        USE_SUDO=true
    fi
}

verify_project_structure() {
    echo -e "\n${CYAN}[INFO]${NC} Verifying project structure..."
    
    # Change to script directory
    cd "$SCRIPT_DIR"
    
    # Check required files
    REQUIRED_FILES=("main.go" "go.mod" "Makefile" "patterns/secrets.txt")
    for file in "${REQUIRED_FILES[@]}"; do
        if [[ ! -f "$file" ]]; then
            echo -e "${RED}[ERROR]${NC} Required file not found: $file"
            echo -e "${YELLOW}[TIP]${NC} Make sure you're in the CodeHunter directory"
            exit 1
        fi
    done
    
    echo -e "${GREEN}[SUCCESS]${NC} Project structure verified ✓"
}

build_and_install() {
    echo -e "\n${CYAN}[INFO]${NC} Building and installing CodeHunter..."
    
    cd "$SCRIPT_DIR"
    
    if [[ "$USE_MAKE" == true ]]; then
        echo -e "${BLUE}[BUILD]${NC} Using Makefile..."
        
        # Check if we can use sudo for install
        if [[ "$USE_SUDO" == true ]]; then
            make install
        else
            echo -e "${YELLOW}[WARN]${NC} No sudo access - Building only"
            make build
            echo -e "${YELLOW}[MANUAL]${NC} Please manually copy 'codehunter' to your PATH"
        fi
        
    else
        echo -e "${BLUE}[BUILD]${NC} Manual installation..."
        
        # Build binary
        echo -e "${CYAN}[INFO]${NC} Building binary..."
        go build -ldflags="-s -w" -o codehunter main.go
        
        if [[ $? -ne 0 ]]; then
            echo -e "${RED}[ERROR]${NC} Build failed!"
            exit 1
        fi
        
        echo -e "${GREEN}[SUCCESS]${NC} Binary built successfully ✓"
        
        # Install binary
        if [[ "$USE_SUDO" == true ]]; then
            echo -e "${CYAN}[INFO]${NC} Installing binary..."
            sudo cp codehunter /usr/local/bin/
            sudo chmod +x /usr/local/bin/codehunter
            
            # Install patterns
            echo -e "${CYAN}[INFO]${NC} Installing patterns..."
            sudo mkdir -p /usr/share/codehunter/patterns
            sudo cp patterns/* /usr/share/codehunter/patterns/
            
            echo -e "${GREEN}[SUCCESS]${NC} CodeHunter installed system-wide ✓"
        else
            echo -e "${YELLOW}[WARN]${NC} No sudo access - Local installation only"
            echo -e "${CYAN}[INFO]${NC} Binary available at: $(pwd)/codehunter"
            echo -e "${YELLOW}[MANUAL]${NC} Add to PATH: export PATH=\$PATH:$(pwd)"
        fi
    fi
}

verify_installation() {
    echo -e "\n${CYAN}[INFO]${NC} Verifying installation..."
    
    # Check if codehunter is in PATH
    if command -v codehunter &> /dev/null; then
        echo -e "${GREEN}[SUCCESS]${NC} CodeHunter installed in PATH ✓"
        CODEHUNTER_CMD="codehunter"
    elif [[ -f "$SCRIPT_DIR/codehunter" ]]; then
        echo -e "${YELLOW}[INFO]${NC} CodeHunter available locally"
        CODEHUNTER_CMD="$SCRIPT_DIR/codehunter"
    else
        echo -e "${RED}[ERROR]${NC} CodeHunter binary not found"
        exit 1
    fi
    
    # Test basic functionality
    echo -e "${BLUE}[TEST]${NC} Testing basic functionality..."
    if $CODEHUNTER_CMD -b=false -r patterns/api_endpoints.txt < /dev/null 2>/dev/null; then
        echo -e "${GREEN}[SUCCESS]${NC} CodeHunter working correctly! ✓"
    else
        echo -e "${YELLOW}[WARN]${NC} CodeHunter installed but basic test failed"
        echo -e "${CYAN}[INFO]${NC} This might be normal - try running manually"
    fi
}

show_usage() {
    echo -e "\n${BOLD}${GREEN}🎉 INSTALLATION COMPLETE! 🎉${NC}"
    echo ""
    echo -e "${BOLD}${YELLOW}📍 Installation Location:${NC}"
    if command -v codehunter &> /dev/null; then
        echo -e "${GREEN}System-wide:${NC} $(which codehunter)"
    else
        echo -e "${YELLOW}Local:${NC} $SCRIPT_DIR/codehunter"
    fi
    echo ""
    
    echo -e "${BOLD}${YELLOW}🎯 Installation Command Used:${NC}"
    echo -e "${BLUE}git clone https://github.com/Acorzo1983/Codehunter.git && cd Codehunter && chmod +x installer.sh && ./installer.sh${NC}"
    echo ""
    
    echo -e "${BOLD}${YELLOW}🎯 Quick Start:${NC}"
    echo -e "${BLUE}Basic scan:${NC}"
    echo "  codehunter -r secrets.txt -l urls.txt -o found.txt"
    echo ""
    echo -e "${BLUE}Pipe with Bug Bounty tools:${NC}"
    echo "  katana -u tesla.com | codehunter -r api_endpoints.txt"
    echo "  subfinder -d tesla.com | httpx | codehunter -r secrets.txt"
    echo "  waybackurls tesla.com | codehunter -r admin_panels.txt"
    echo ""
    echo -e "${BLUE}With proxychains:${NC}"
    echo "  proxychains codehunter -r secrets.txt -l urls.txt"
    echo ""
    
    echo -e "${BOLD}${PURPLE}📋 Available Patterns:${NC}"
    echo "  • secrets.txt      - API keys, tokens, credentials"
    echo "  • api_endpoints.txt - REST APIs, endpoints"
    echo "  • admin_panels.txt  - Admin areas, panels"
    echo "  • js_secrets.txt   - JavaScript secrets"
    echo "  • files.txt        - Sensitive files"
    echo "  • custom.txt       - Your custom patterns"
    echo ""
    
    echo -e "${BOLD}${CYAN}📁 Pattern Files Location:${NC}"
    if [[ -d "/usr/share/codehunter/patterns" ]]; then
        echo -e "${GREEN}System:${NC} /usr/share/codehunter/patterns/"
    fi
    echo -e "${BLUE}Local:${NC} $SCRIPT_DIR/patterns/"
    echo ""
    
    echo -e "${BOLD}${GREEN}🧪 Test Installation:${NC}"
    echo "  codehunter -r patterns/api_endpoints.txt -l examples/urls.txt -v"
    echo ""
    
    echo -e "${BOLD}${CYAN}🏴‍☠️ Made with ❤️ by Albert.C @yz9yt 🏴‍☠️${NC}"
    echo -e "${PURPLE}GitHub: https://github.com/Acorzo1983/Codehunter${NC}"
    echo -e "${PURPLE}Twitter: @yz9yt${NC}"
    echo ""
    echo -e "${BOLD}🎯 Happy Bug Hunting! 🎯${NC}"
}

cleanup_on_error() {
    echo -e "\n${RED}[ERROR]${NC} Installation failed!"
    echo -e "${CYAN}[INFO]${NC} Cleaning up..."
    
    # Remove any partially installed files
    [[ -f "codehunter" ]] && rm -f codehunter
    
    echo -e "${YELLOW}[HELP]${NC} Try manual installation:"
    echo "  go build -o codehunter main.go"
    echo "  sudo cp codehunter /usr/local/bin/"
    echo ""
    echo -e "${PURPLE}Need help? https://github.com/Acorzo1983/Codehunter/issues${NC}"
}

show_help() {
    echo -e "${BOLD}${PURPLE}CodeHunter v2.5 Installer Help${NC}"
    echo ""
    echo -e "${YELLOW}Usage:${NC}"
    echo "  ./installer.sh [options]"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo "  -h, --help     Show this help message"
    echo "  -v, --verbose  Verbose output"
    echo "  --no-sudo      Don't use sudo (local install only)"
    echo "  --force        Force installation even if already installed"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  ./installer.sh"
    echo "  ./installer.sh --verbose"
    echo "  ./installer.sh --no-sudo"
    echo ""
    echo -e "${CYAN}Made with ❤️ by Albert.C @yz9yt${NC}"
}

# Parse command line arguments
VERBOSE=false
FORCE_INSTALL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --no-sudo)
            USE_SUDO=false
            shift
            ;;
        --force)
            FORCE_INSTALL=true
            shift
            ;;
        *)
            echo -e "${RED}[ERROR]${NC} Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Main installation function
main() {
    # Set verbose output
    if [[ "$VERBOSE" == true ]]; then
        set -x
    fi
    
    print_banner
    
    # Check if already installed
    if command -v codehunter &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
        echo -e "${YELLOW}[INFO]${NC} CodeHunter is already installed: $(which codehunter)"
        echo -e "${CYAN}[INFO]${NC} Use --force to reinstall"
        read -p "Continue with reinstallation? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${GREEN}[INFO]${NC} Installation cancelled"
            exit 0
        fi
    fi
    
    check_os
    check_dependencies
    verify_project_structure
    build_and_install
    verify_installation
    show_usage
    
    echo -e "${GREEN}[SUCCESS]${NC} Installation completed successfully! 🎉"
}

# Handle interrupts and errors
trap cleanup_on_error EXIT

# Store original directory
ORIGINAL_DIR="$CURRENT_DIR"

# Ensure we return to original directory on exit
cleanup() {
    cd "$ORIGINAL_DIR" 2>/dev/null || true
}
trap cleanup EXIT

# Run main function
main "$@"

# If we get here, installation was successful
trap - EXIT
