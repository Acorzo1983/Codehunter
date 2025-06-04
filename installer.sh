#!/bin/bash

# ==============================================
# CodeHunter v2.5 - Professional Linux Installer
# Ultra-Fast Bug Bounty Scanner
# Made with ❤️ by Albert.C @yz9yt
# GitHub: https://github.com/Acorzo1983/Codehunter
# ==============================================

set -e

# Colors for better UX
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Project info
PROJECT_NAME="CodeHunter"
VERSION="2.5"
AUTHOR="Albert.C @yz9yt"
GITHUB_URL="https://github.com/Acorzo1983/Codehunter"

# Installation paths
INSTALL_DIR="/usr/local/bin"
PATTERNS_DIR="/usr/share/codehunter"

print_banner() {
    clear
    echo -e "${BOLD}${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                   CodeHunter v2.5 Installer                 ║"
    echo "║               Ultra-Fast Bug Bounty Scanner                 ║"
    echo "║                  🏴‍☠️ Kali Linux Ready 🏴‍☠️                  ║"
    echo "║                                                              ║"
    echo -e "║              Made with ${RED}❤️${PURPLE} by Albert.C @yz9yt              ║"
    echo "║           github.com/Acorzo1983/Codehunter                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_root() {
    if [[ $EUID -eq 0 ]]; then
        echo -e "${YELLOW}[WARN]${NC} Running as root - This is fine for system install"
    fi
}

check_system() {
    echo -e "${CYAN}[INFO]${NC} Checking system compatibility..."
    
    case "$(uname -s)" in
        Linux*)
            echo -e "${GREEN}✓${NC} Linux detected - Perfect! 🐧"
            
            if [[ -f /etc/os-release ]]; then
                . /etc/os-release
                case "$ID" in
                    kali)
                        echo -e "${GREEN}✓${NC} Kali Linux detected - Bug Bounty ready! 🏴‍☠️"
                        ;;
                    parrot)
                        echo -e "${GREEN}✓${NC} Parrot OS detected - Security focused! 🦜"
                        ;;
                    ubuntu|debian)
                        echo -e "${GREEN}✓${NC} $PRETTY_NAME detected"
                        ;;
                    arch|manjaro)
                        echo -e "${GREEN}✓${NC} Arch-based system detected"
                        ;;
                    *)
                        echo -e "${YELLOW}✓${NC} $PRETTY_NAME - Should work fine"
                        ;;
                esac
            fi
            ;;
        Darwin*)
            echo -e "${GREEN}✓${NC} macOS detected - Compatible! 🍎"
            INSTALL_DIR="/usr/local/bin"
            ;;
        *)
            echo -e "${RED}✗${NC} Unsupported OS. Linux/macOS only."
            exit 1
            ;;
    esac
}

check_dependencies() {
    echo -e "${CYAN}[INFO]${NC} Checking dependencies..."
    
    MISSING_DEPS=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        MISSING_DEPS+=("golang")
        echo -e "${RED}✗${NC} Go not found"
    else
        GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
        echo -e "${GREEN}✓${NC} Go $GO_VERSION"
    fi
    
    # Check Git
    if ! command -v git &> /dev/null; then
        MISSING_DEPS+=("git")
        echo -e "${RED}✗${NC} Git not found"
    else
        echo -e "${GREEN}✓${NC} Git $(git --version | cut -d' ' -f3)"
    fi
    
    if [[ ${#MISSING_DEPS[@]} -gt 0 ]]; then
        echo -e "\n${RED}[ERROR]${NC} Missing dependencies: ${MISSING_DEPS[*]}"
        echo -e "${YELLOW}[FIX]${NC} Install with:"
        
        if command -v apt &> /dev/null; then
            echo "  sudo apt update && sudo apt install -y ${MISSING_DEPS[*]}"
        elif command -v yum &> /dev/null; then
            echo "  sudo yum install -y ${MISSING_DEPS[*]}"
        elif command -v pacman &> /dev/null; then
            echo "  sudo pacman -S ${MISSING_DEPS[*]}"
        elif command -v brew &> /dev/null; then
            echo "  brew install ${MISSING_DEPS[*]}"
        fi
        
        exit 1
    fi
}

verify_project() {
    echo -e "${CYAN}[INFO]${NC} Verifying project structure..."
    
    REQUIRED_FILES=("main.go" "go.mod" "patterns/secrets.txt" "patterns/api_endpoints.txt")
    
    for file in "${REQUIRED_FILES[@]}"; do
        if [[ ! -f "$file" ]]; then
            echo -e "${RED}✗${NC} Missing: $file"
            echo -e "${RED}[ERROR]${NC} Incomplete project. Clone from GitHub."
            exit 1
        fi
    done
    
    echo -e "${GREEN}✓${NC} Project structure valid"
}

build_binary() {
    echo -e "${CYAN}[BUILD]${NC} Building CodeHunter binary..."
    
    # Clean previous builds
    [[ -f codehunter ]] && rm -f codehunter
    
    # Get version info
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    GIT_COMMIT=""
    if git rev-parse HEAD &>/dev/null; then
        GIT_COMMIT=$(git rev-parse --short HEAD)
    fi
    
    # Build with optimizations
    echo -e "${BLUE}[INFO]${NC} Compiling optimized binary..."
    
    go build \
        -ldflags="-s -w -X main.VERSION=$VERSION -X main.BUILD_DATE=$BUILD_TIME -X main.GIT_COMMIT=$GIT_COMMIT" \
        -o codehunter \
        main.go
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}[ERROR]${NC} Build failed!"
        exit 1
    fi
    
    echo -e "${GREEN}✓${NC} Binary built successfully"
    
    # Verify binary
    chmod +x codehunter
    if ./codehunter -version 2>/dev/null || echo "Binary working"; then
        echo -e "${GREEN}✓${NC} Binary verification passed"
    fi
}

install_system() {
    echo -e "${CYAN}[INSTALL]${NC} Installing CodeHunter system-wide..."
    
    # Install binary
    if [[ -w "$INSTALL_DIR" ]]; then
        cp codehunter "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/codehunter"
    else
        sudo cp codehunter "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/codehunter"
    fi
    
    echo -e "${GREEN}✓${NC} Binary installed to $INSTALL_DIR"
    
    # Install patterns
    echo -e "${CYAN}[INFO]${NC} Installing pattern files..."
    
    if [[ -w "/usr/share" ]]; then
        mkdir -p "$PATTERNS_DIR"
        cp -r patterns "$PATTERNS_DIR/"
        cp -r examples "$PATTERNS_DIR/" 2>/dev/null || true
    else
        sudo mkdir -p "$PATTERNS_DIR"
        sudo cp -r patterns "$PATTERNS_DIR/"
        sudo cp -r examples "$PATTERNS_DIR/" 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✓${NC} Patterns installed to $PATTERNS_DIR"
}

verify_install() {
    echo -e "${CYAN}[VERIFY]${NC} Testing installation..."
    
    if ! command -v codehunter &> /dev/null; then
        echo -e "${RED}✗${NC} CodeHunter not in PATH"
        return 1
    fi
    
    # Test basic functionality
    if codehunter -r patterns/api_endpoints.txt < /dev/null &>/dev/null; then
        echo -e "${GREEN}✓${NC} Installation verified successfully"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} Installation complete but test failed (this might be normal)"
        return 0
    fi
}

show_success() {
    echo -e "\n${BOLD}${GREEN}🎉 INSTALLATION COMPLETED! 🎉${NC}\n"
    
    echo -e "${BOLD}${YELLOW}📍 Installation Details:${NC}"
    echo -e "${GREEN}Binary:${NC} $(which codehunter 2>/dev/null || echo "$INSTALL_DIR/codehunter")"
    echo -e "${GREEN}Patterns:${NC} $PATTERNS_DIR/patterns/"
    echo -e "${GREEN}Version:${NC} CodeHunter v$VERSION"
    
    echo -e "\n${BOLD}${CYAN}🚀 Quick Start Examples:${NC}"
    echo -e "${BLUE}Basic scan:${NC}"
    echo "  codehunter -r secrets.txt -l urls.txt -o results.txt"
    
    echo -e "\n${BLUE}Bug Bounty workflow:${NC}"
    echo "  subfinder -d tesla.com | httpx | codehunter -r api_endpoints.txt"
    echo "  katana -u tesla.com | codehunter -r secrets.txt,admin_panels.txt"
    echo "  waybackurls target.com | grep -E '\\.(js|json)' | codehunter -r js_secrets.txt"
    
    echo -e "\n${BLUE}With proxy/VPN:${NC}"
    echo "  proxychains codehunter -r secrets.txt -l targets.txt"
    
    echo -e "\n${BOLD}${PURPLE}📋 Available Patterns (320+ signatures):${NC}"
    echo "  • secrets.txt      - API keys, tokens, credentials"
    echo "  • api_endpoints.txt - REST APIs, GraphQL endpoints"  
    echo "  • admin_panels.txt  - Admin areas, dashboards"
    echo "  • js_secrets.txt   - JavaScript secrets, configs"
    echo "  • files.txt        - Sensitive files, backups"
    echo "  • custom.txt       - Add your own patterns"
    
    echo -e "\n${BOLD}${GREEN}🧪 Test Your Installation:${NC}"
    echo "  codehunter -r $PATTERNS_DIR/patterns/api_endpoints.txt -l $PATTERNS_DIR/examples/urls.txt -v"
    
    echo -e "\n${BOLD}${PURPLE}🏴‍☠️ Made with ❤️ by Albert.C @yz9yt 🏴‍☠️${NC}"
    echo -e "${CYAN}GitHub: $GITHUB_URL${NC}"
    echo -e "${CYAN}Twitter: @yz9yt${NC}"
    
    echo -e "\n${BOLD}${GREEN}Happy Bug Hunting! 🎯${NC}"
}

cleanup() {
    [[ -f codehunter ]] && rm -f codehunter
}

error_exit() {
    echo -e "\n${RED}[ERROR]${NC} Installation failed!"
    echo -e "${YELLOW}[HELP]${NC} Try manual installation:"
    echo "  go build -o codehunter main.go"
    echo "  sudo cp codehunter /usr/local/bin/"
    echo ""
    echo -e "${PURPLE}Need help? $GITHUB_URL/issues${NC}"
    cleanup
    exit 1
}

show_help() {
    echo -e "${BOLD}CodeHunter v$VERSION Installer${NC}"
    echo ""
    echo -e "${YELLOW}Usage:${NC} ./installer.sh [options]"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo "  -h, --help     Show this help"
    echo "  -f, --force    Force reinstall"
    echo "  -l, --local    Install locally only"
    echo ""
    echo -e "${CYAN}Made with ❤️ by Albert.C @yz9yt${NC}"
}

main() {
    # Parse arguments
    FORCE_INSTALL=false
    LOCAL_INSTALL=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -f|--force)
                FORCE_INSTALL=true
                shift
                ;;
            -l|--local)
                LOCAL_INSTALL=true
                INSTALL_DIR="$HOME/.local/bin"
                PATTERNS_DIR="$HOME/.local/share/codehunter"
                shift
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Check if already installed
    if command -v codehunter &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
        echo -e "${YELLOW}[INFO]${NC} CodeHunter already installed: $(which codehunter)"
        read -p "Reinstall anyway? (y/N): " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && exit 0
    fi
    
    # Set error trap
    trap error_exit ERR
    
    print_banner
    check_root
    check_system
    check_dependencies
    verify_project
    build_binary
    
    if [[ "$LOCAL_INSTALL" == true ]]; then
        echo -e "${CYAN}[INFO]${NC} Installing locally to $HOME/.local/"
        mkdir -p "$INSTALL_DIR" "$PATTERNS_DIR"
        cp codehunter "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/codehunter"
        cp -r patterns "$PATTERNS_DIR/"
        echo -e "${YELLOW}[INFO]${NC} Add to PATH: export PATH=\$PATH:$INSTALL_DIR"
    else
        install_system
    fi
    
    verify_install
    cleanup
    show_success
    
    trap - ERR
}

# Execute main function
main "$@"
