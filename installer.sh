#!/bin/bash
# CodeHunter v2.5 - Linux Native Installer
# Made with ❤️ by Albert.C @yz9yt
# https://github.com/Acorzo1983/Codehunter

set -e

# Colors
R='\033[0;31m'
G='\033[0;32m'
Y='\033[1;33m'
B='\033[0;34m'
P='\033[0;35m'
C='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

VERSION="2.5"
INSTALL_DIR="/usr/local/bin"
PATTERNS_DIR="/usr/share/codehunter"

banner() {
    clear
    echo -e "${BOLD}${P}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                   CodeHunter v2.5 Installer                 ║"
    echo "║               Ultra-Fast Bug Bounty Scanner                 ║"
    echo "║                  🏴‍☠️ Kali Linux Ready 🏴‍☠️                  ║"
    echo "║                                                              ║"
    echo -e "║              Made with ${R}❤️${P} by Albert.C @yz9yt              ║"
    echo "║           github.com/Acorzo1983/Codehunter                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

check_os() {
    echo -e "${C}[INFO]${NC} Checking system..."
    case "$(uname -s)" in
        Linux*) 
            echo -e "${G}✓${NC} Linux detected 🐧"
            if [[ -f /etc/os-release ]]; then
                . /etc/os-release
                echo -e "${G}✓${NC} Running on $PRETTY_NAME"
            fi
            ;;
        Darwin*) 
            echo -e "${G}✓${NC} macOS detected 🍎"
            ;;
        *) 
            echo -e "${R}✗${NC} Unsupported OS"
            exit 1
            ;;
    esac
}

check_deps() {
    echo -e "${C}[INFO]${NC} Checking dependencies..."

    if ! command -v go &>/dev/null; then
        echo -e "${R}✗${NC} Go not found"
        echo -e "${Y}[FIX]${NC} Install Go: sudo apt install golang-go"
        exit 1
    fi
    echo -e "${G}✓${NC} Go $(go version | awk '{print $3}')"

    if ! command -v git &>/dev/null; then
        echo -e "${R}✗${NC} Git not found"
        exit 1
    fi
    echo -e "${G}✓${NC} Git found"
}

verify_files() {
    echo -e "${C}[INFO]${NC} Verifying project..."

    for file in "main.go" "go.mod" "patterns/secrets.txt"; do
        if [[ ! -f "$file" ]]; then
            echo -e "${R}✗${NC} Missing: $file"
            exit 1
        fi
    done
    echo -e "${G}✓${NC} Project structure valid"
}

build() {
    echo -e "${C}[BUILD]${NC} Building CodeHunter..."

    [[ -f codehunter ]] && rm -f codehunter

    go build -ldflags="-s -w" -o codehunter main.go

    chmod +x codehunter
    echo -e "${G}✓${NC} Binary built successfully"
}

install() {
    echo -e "${C}[INSTALL]${NC} Installing CodeHunter..."

    if [[ -w "$INSTALL_DIR" ]]; then
        cp codehunter "$INSTALL_DIR/"
    else
        sudo cp codehunter "$INSTALL_DIR/"
    fi

    echo -e "${G}✓${NC} Binary installed to $INSTALL_DIR"

    if [[ -w "/usr/share" ]]; then
        mkdir -p "$PATTERNS_DIR"
        cp -r patterns "$PATTERNS_DIR/"
    else
        sudo mkdir -p "$PATTERNS_DIR"
        sudo cp -r patterns "$PATTERNS_DIR/"
    fi

    echo -e "${G}✓${NC} Patterns installed to $PATTERNS_DIR"
}

verify() {
    echo -e "${C}[VERIFY]${NC} Testing installation..."

    if command -v codehunter &>/dev/null; then
        echo -e "${G}✓${NC} Installation verified"
    else
        echo -e "${R}✗${NC} Installation failed"
        exit 1
    fi
}

success() {
    echo -e "\n${BOLD}${G}🎉 INSTALLATION COMPLETE! 🎉${NC}\n"

    echo -e "${BOLD}${Y}📍 Installation Details:${NC}"
    echo -e "${G}Binary:${NC} $(which codehunter)"
    echo -e "${G}Patterns:${NC} $PATTERNS_DIR/patterns/"

    echo -e "\n${BOLD}${C}🚀 Quick Start:${NC}"
    echo -e "${B}Basic scan:${NC}"
    echo "  codehunter -r secrets.txt -l urls.txt -o results.txt"

    echo -e "\n${B}Bug Bounty workflow:${NC}"
    echo "  subfinder -d tesla.com | httpx | codehunter -r api_endpoints.txt"
    echo "  katana -u tesla.com | codehunter -r secrets.txt"
    echo "  waybackurls target.com | codehunter -r js_secrets.txt"

    echo -e "\n${BOLD}${P}📋 Available Patterns:${NC}"
    echo "  • secrets.txt      - API keys, tokens"
    echo "  • api_endpoints.txt - REST APIs"
    echo "  • admin_panels.txt  - Admin areas"
    echo "  • js_secrets.txt   - JavaScript secrets"
    echo "  • files.txt        - Sensitive files"

    echo -e "\n${BOLD}${G}🧪 Test Installation:${NC}"
    echo "  codehunter -r $PATTERNS_DIR/patterns/api_endpoints.txt -v"

    echo -e "\n${BOLD}${P}🏴‍☠️ Made with ❤️ by Albert.C @yz9yt 🏴‍☠️${NC}"
    echo -e "${C}GitHub: https://github.com/Acorzo1983/Codehunter${NC}"

    echo -e "\n${BOLD}${G}Happy Bug Hunting! 🎯${NC}"
}

cleanup() {
    [[ -f codehunter ]] && rm -f codehunter
}

error_exit() {
    echo -e "\n${R}[ERROR]${NC} Installation failed!"
    cleanup
    exit 1
}

help() {
    echo -e "${BOLD}CodeHunter v$VERSION Installer${NC}"
    echo ""
    echo -e "${Y}Usage:${NC} ./installer.sh [options]"
    echo ""
    echo -e "${Y}Options:${NC}"
    echo "  -h, --help     Show help"
    echo "  -l, --local    Local install only"
    echo ""
    echo -e "${C}Made with ❤️ by Albert.C @yz9yt${NC}"
}

main() {
    LOCAL_INSTALL=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                help
                exit 0
                ;;
            -l|--local)
                LOCAL_INSTALL=true
                INSTALL_DIR="$HOME/.local/bin"
                PATTERNS_DIR="$HOME/.local/share/codehunter"
                shift
                ;;
            *)
                echo -e "${R}Unknown option: $1${NC}"
                exit 1
                ;;
        esac
    done

    trap error_exit ERR

    banner
    check_os
    check_deps
    verify_files
    build

    if [[ "$LOCAL_INSTALL" == true ]]; then
        echo -e "${C}[INFO]${NC} Installing locally..."
        mkdir -p "$INSTALL_DIR" "$PATTERNS_DIR"
        cp codehunter "$INSTALL_DIR/"
        cp -r patterns "$PATTERNS_DIR/"
        echo -e "${Y}[INFO]${NC} Add to PATH: export PATH=\$PATH:$INSTALL_DIR"
    else
        install
    fi

    verify
    cleanup
    success

    trap - ERR
}

main "$@"
