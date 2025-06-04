#!/bin/bash

# ╔══════════════════════════════════════════════════════╗
# ║                CodeHunter Installer v2.5            ║
# ║          Ultra-Fast Bug Bounty Scanner              ║
# ║                                                      ║
# ║           Made with ❤️ by Albert.C @yz9yt           ║
# ║        https://github.com/Acorzo1983/Codehunter     ║
# ╚══════════════════════════════════════════════════════╝

set -e

VERSION="2.5"
AUTHOR="Albert.C @yz9yt"
GITHUB_REPO="https://github.com/Acorzo1983/Codehunter"
GITHUB_RAW="https://raw.githubusercontent.com/Acorzo1983/Codehunter/main"
GITHUB_API="https://api.github.com/repos/Acorzo1983/Codehunter"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.codehunter"
PATTERNS_DIR="/usr/share/codehunter/patterns"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

print_banner() {
    echo -e "${BOLD}${CYAN}"
    echo "   ╔══════════════════════════════════════════════════════╗"
    echo "   ║                CodeHunter v${VERSION} Installer            ║"
    echo "   ║          Ultra-Fast Bug Bounty Scanner              ║"
    echo "   ║                                                      ║"
    echo -e "   ║           Made with ${RED}❤️${CYAN} by Albert.C @yz9yt           ║"
    echo "   ║        github.com/Acorzo1983/Codehunter             ║"
    echo "   ╚══════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

detect_os() {
    case "$(uname -s)" in
        Linux*)     OS=linux;;
        Darwin*)    OS=darwin;;
        MINGW*|MSYS*|CYGWIN*) OS=windows;;
        *)          OS=unknown;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   ARCH=amd64;;
        arm64|aarch64)  ARCH=arm64;;
        armv7l)         ARCH=arm;;
        i386|i686)      ARCH=386;;
        *)              ARCH=unknown;;
    esac
    
    echo -e "${BLUE}[INFO]${NC} Detected OS: ${OS}, Architecture: ${ARCH}"
}

check_requirements() {
    echo -e "${YELLOW}[INFO]${NC} Checking requirements..."
    
    # Check internet connectivity
    if ! curl -s --head ${GITHUB_REPO} > /dev/null; then
        echo -e "${RED}[ERROR]${NC} Cannot connect to GitHub!"
        echo "Please check your internet connection"
        exit 1
    fi
    echo -e "${GREEN}[INFO]${NC} GitHub connectivity ✓"
    
    # Check required tools
    for tool in curl; do
        if ! command -v $tool &> /dev/null; then
            echo -e "${RED}[ERROR]${NC} $tool is required but not installed!"
            echo "Please install $tool and try again"
            exit 1
        fi
    done
    echo -e "${GREEN}[INFO]${NC} Required tools found ✓"
    
    # Check permissions
    if [[ $EUID -eq 0 ]]; then
        SUDO=""
    else
        SUDO="sudo"
        echo -e "${YELLOW}[INFO]${NC} Will use sudo for system installation"
    fi
}

get_latest_release() {
    echo -e "${YELLOW}[INFO]${NC} Getting latest release information..."
    
    # Try to get the latest release
    LATEST_RELEASE=$(curl -s "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null || echo "")
    
    if [[ -n "$LATEST_RELEASE" ]]; then
        VERSION="$LATEST_RELEASE"
        echo -e "${GREEN}[INFO]${NC} Latest release: $VERSION ✓"
    else
        echo -e "${YELLOW}[WARN]${NC} Could not fetch latest release, using default version $VERSION"
    fi
}

download_prebuilt_binary() {
    echo -e "${YELLOW}[INFO]${NC} Attempting to download prebuilt binary..."
    
    # Construct download URL
    BINARY_NAME="codehunter"
    if [[ "$OS" == "windows" ]]; then
        BINARY_NAME="codehunter.exe"
    fi
    
    # Try different release asset naming conventions
    POSSIBLE_NAMES=(
        "codehunter-${OS}-${ARCH}"
        "codehunter-${VERSION}-${OS}-${ARCH}"
        "codehunter_${OS}_${ARCH}"
    )
    
    for NAME in "${POSSIBLE_NAMES[@]}"; do
        DOWNLOAD_URL="${GITHUB_REPO}/releases/download/${VERSION}/${NAME}"
        if [[ "$OS" == "windows" ]]; then
            DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
        fi
        
        echo -e "${CYAN}[INFO]${NC} Trying: $DOWNLOAD_URL"
        
        if curl -sL --fail "$DOWNLOAD_URL" -o "/tmp/codehunter" 2>/dev/null; then
            chmod +x "/tmp/codehunter"
            if [[ -x "/tmp/codehunter" ]]; then
                echo -e "${GREEN}[INFO]${NC} Prebuilt binary downloaded successfully ✓"
                return 0
            fi
        fi
    done
    
    echo -e "${YELLOW}[WARN]${NC} No prebuilt binary found, will build from source"
    return 1
}

build_from_source() {
    echo -e "${YELLOW}[INFO]${NC} Building CodeHunter from source..."
    echo -e "${PURPLE}[INFO]${NC} This requires Go to be installed"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}[ERROR]${NC} Go is not installed!"
        echo "Please install Go from: https://golang.org/dl/"
        echo "Or try the prebuilt binary option"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}[INFO]${NC} Go ${GO_VERSION} found ✓"
    
    # Check if git is available for cloning
    if command -v git &> /dev/null; then
        TEMP_DIR=$(mktemp -d)
        cd "$TEMP_DIR"
        
        echo -e "${CYAN}[INFO]${NC} Cloning repository..."
        git clone --depth=1 "${GITHUB_REPO}" . || {
            echo -e "${RED}[ERROR]${NC} Failed to clone repository!"
            exit 1
        }
        
        # Use Makefile if available
        if [[ -f "Makefile" ]]; then
            echo -e "${GREEN}[INFO]${NC} Using Makefile for professional build ✨"
            make build || {
                echo -e "${RED}[ERROR]${NC} Build failed using Makefile!"
                exit 1
            }
            cp build/codehunter /tmp/codehunter
        else
            # Fallback to manual build
            echo -e "${YELLOW}[INFO]${NC} Building manually..."
            go mod init codehunter 2>/dev/null || true
            go mod tidy 2>/dev/null || true
            go build -ldflags="-s -w -X 'main.VERSION=${VERSION}' -X 'main.AUTHOR=${AUTHOR}'" -o /tmp/codehunter . || {
                echo -e "${RED}[ERROR]${NC} Manual build failed!"
                exit 1
            }
        fi
    else
        # Download source and build
        echo -e "${CYAN}[INFO]${NC} Downloading source code..."
        TEMP_DIR=$(mktemp -d)
        cd "$TEMP_DIR"
        
        curl -sSL "${GITHUB_RAW}/main.go" -o main.go || {
            echo -e "${RED}[ERROR]${NC} Failed to download source code!"
            exit 1
        }
        
        # Try to download go.mod
        curl -sSL "${GITHUB_RAW}/go.mod" -o go.mod 2>/dev/null || {
            echo -e "${YELLOW}[INFO]${NC} Creating go.mod..."
            go mod init codehunter
        }
        
        go mod tidy 2>/dev/null || true
        go build -ldflags="-s -w -X 'main.VERSION=${VERSION}' -X 'main.AUTHOR=${AUTHOR}'" -o /tmp/codehunter . || {
            echo -e "${RED}[ERROR]${NC} Build failed!"
            exit 1
        }
    fi
    
    chmod +x /tmp/codehunter
    echo -e "${GREEN}[INFO]${NC} Build completed successfully ✓"
}

install_binary() {
    echo -e "${YELLOW}[INFO]${NC} Installing CodeHunter to ${INSTALL_DIR}..."
    
    # Verify binary works
    if ! /tmp/codehunter -version >/dev/null 2>&1; then
        echo -e "${RED}[ERROR]${NC} Binary verification failed!"
        exit 1
    fi
    
    $SUDO cp /tmp/codehunter "$INSTALL_DIR/codehunter"
    $SUDO chmod +x "$INSTALL_DIR/codehunter"
    
    # Clean up temp file
    rm -f /tmp/codehunter
    
    echo -e "${GREEN}[INFO]${NC} Binary installed ✓"
    echo -e "${CYAN}[INFO]${NC} Ready to hunt bugs with tools by ${AUTHOR}!"
}

create_directories() {
    echo -e "${YELLOW}[INFO]${NC} Creating directories..."
    
    # User config directory
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$CONFIG_DIR/examples"
    
    # System patterns directory
    $SUDO mkdir -p "$PATTERNS_DIR"
    
    echo -e "${GREEN}[INFO]${NC} Directories created ✓"
}

install_patterns() {
    echo -e "${YELLOW}[INFO]${NC} Installing curated pattern files..."
    echo -e "${PURPLE}[INFO]${NC} These patterns were handcrafted by ${AUTHOR}"
    
    # Function to download patterns from GitHub
    download_pattern() {
        local pattern_file="$1"
        echo -e "${CYAN}[INFO]${NC} Downloading ${pattern_file}..."
        
        if curl -sSL "${GITHUB_RAW}/patterns/${pattern_file}" -o "/tmp/${pattern_file}" 2>/dev/null; then
            $SUDO cp "/tmp/${pattern_file}" "$PATTERNS_DIR/"
            rm -f "/tmp/${pattern_file}"
            echo -e "${GREEN}[INFO]${NC} ${pattern_file} installed ✓"
            return 0
        else
            echo -e "${YELLOW}[WARN]${NC} Could not download ${pattern_file}, creating default..."
            return 1
        fi
    }
    
    # Create header for pattern files
    PATTERN_HEADER="# CodeHunter v${VERSION} Pattern File
# Made with ❤️ by ${AUTHOR}
# GitHub: ${GITHUB_REPO}
# Twitter: @yz9yt
# 
# This file contains carefully curated patterns for bug bounty hunting
#"

    # Download or create pattern files
    declare -A PATTERNS=(
        ["secrets.txt"]="API Keys and Secrets"
        ["api_endpoints.txt"]="API Endpoints and Paths"
        ["js_secrets.txt"]="JavaScript Vulnerabilities"
        ["custom.txt"]="Custom Patterns"
    )
    
    for pattern_file in "${!PATTERNS[@]}"; do
        if ! download_pattern "$pattern_file"; then
            create_default_pattern "$pattern_file" "${PATTERNS[$pattern_file]}"
        fi
    done
    
    echo -e "${GREEN}[INFO]${NC} All pattern files installed ✓"
}

create_default_pattern() {
    local file="$1"
    local description="$2"
    
    case "$file" in
        "secrets.txt")
            $SUDO tee "$PATTERNS_DIR/secrets.txt" > /dev/null << EOF
$PATTERN_HEADER

# API Keys and Secrets
password
passwd
secret
token
api_key
apikey
api-key
access_key
secret_key
private_key
auth_token
bearer
jwt

# Cloud Provider Keys
AKIA[0-9A-Z]{16}
ya29\.[0-9A-Za-z\-_]+
ghp_[0-9a-zA-Z]{36}
sk-[0-9a-zA-Z]{48}
xoxb-[0-9]{10,12}-[0-9]{10,12}-[0-9a-zA-Z]{24}

# Database connections
mongodb://
mysql://
postgres://
redis://
jdbc:
connectionString

# Common vulnerabilities
eval\(
innerHTML
document\.write
\.execute\(
system\(
shell_exec
passthru
EOF
            ;;
        "api_endpoints.txt")
            $SUDO tee "$PATTERNS_DIR/api_endpoints.txt" > /dev/null << EOF
$PATTERN_HEADER

# API Endpoints
/api/
/rest/
/graphql
/swagger
/v1/
/v2/
/v3/
\.json
/admin/api/
/internal/
/private/
/debug/
/test/
/dev/
/staging/
EOF
            ;;
        "js_secrets.txt")
            $SUDO tee "$PATTERNS_DIR/js_secrets.txt" > /dev/null << EOF
$PATTERN_HEADER

# JavaScript Specific Patterns
console\.log
debugger
localhost
127\.0\.0\.1
\.local
staging
development
test
debug
admin
root
password
token
secret
key
config
var\s+.*=.*["\'][^"\']*["\']
let\s+.*=.*["\'][^"\']*["\']
const\s+.*=.*["\'][^"\']*["\']
EOF
            ;;
        "custom.txt")
            $SUDO tee "$PATTERNS_DIR/custom.txt" > /dev/null << EOF
$PATTERN_HEADER

# Albert's Special Custom Patterns
# Handpicked from real-world bug bounty experience

custom-cookie
custom_cookie
customcookie
custom-cookies
custom_cookies
customcookies

# Add your own patterns here
# Made with ❤️ by ${AUTHOR}
EOF
            ;;
    esac
    
    echo -e "${GREEN}[INFO]${NC} Created default ${file} ✓"
}

install_examples() {
    echo -e "${YELLOW}[INFO]${NC} Installing examples and documentation..."
    
    # Download examples or create defaults
    curl -sSL "${GITHUB_RAW}/examples/urls.txt" -o "$CONFIG_DIR/examples/urls.txt" 2>/dev/null || {
        cat > "$CONFIG_DIR/examples/urls.txt" << 'EOF'
# Example URLs for CodeHunter
# Made with ❤️ by Albert.C @yz9yt

https://example.com
https://example.com/api/config
https://example.com/js/app.js
https://example.com/admin/
EOF
    }
    
    # Create comprehensive demo script
    cat > "$CONFIG_DIR/examples/demo.sh" << EOF
#!/bin/bash

# ╔══════════════════════════════════════════════════════╗
# ║                CodeHunter v${VERSION} Demo                  ║
# ║           Made with ❤️ by Albert.C @yz9yt           ║
# ║        github.com/Acorzo1983/Codehunter             ║
# ╚══════════════════════════════════════════════════════╝

echo "🎯 CodeHunter Usage Examples:"
echo "============================="
echo ""

echo "📁 Basic Usage:"
echo "  codehunter -f examples/urls.txt -r secrets.txt -v"
echo "  codehunter -f urls.txt -r api_endpoints.txt -o results.json -format json"
echo ""

echo "🔍 Advanced Scanning:"
echo "  codehunter -f urls.txt -r js_secrets.txt -js-only -show-matches"
echo "  codehunter -f large_urls.txt -r custom.txt -w 50 -v"
echo ""

echo "🔧 Tool Integration:"
echo "  katana -u target.com | codehunter -f stdin -r secrets.txt"
echo "  waybackurls target.com | grep -E '\.(js|json)$' | codehunter -f stdin -r js_secrets.txt"
echo "  subfinder -d target.com | httpx | codehunter -f stdin -r api_endpoints.txt"
echo ""

echo "📊 Output Formats:"
echo "  codehunter -f urls.txt -r patterns.txt -o results.txt -format txt"
echo "  codehunter -f urls.txt -r patterns.txt -o results.csv -format csv"
echo "  codehunter -f urls.txt -r patterns.txt -o results.json -format json"
echo ""

echo "🎯 Bug Bounty Workflow:"
echo "  # 1. Subdomain discovery"
echo "  subfinder -d target.com | httpx -mc 200,301,302 > live_urls.txt"
echo "  "
echo "  # 2. URL discovery"
echo "  cat live_urls.txt | katana -d 3 > all_urls.txt"
echo "  "
echo "  # 3. Pattern hunting"
echo "  codehunter -f all_urls.txt -r secrets.txt -o critical_findings.json -v"
echo ""

echo "🔗 Links:"
echo "========="
echo "GitHub: ${GITHUB_REPO}"
echo "Made with ❤️ by ${AUTHOR}"
echo "Follow @yz9yt for more security tools!"
echo ""
echo "🚀 Happy Bug Hunting!"
EOF

    chmod +x "$CONFIG_DIR/examples/demo.sh"
    
    # Download or create README
    curl -sSL "${GITHUB_RAW}/README.md" -o "$CONFIG_DIR/README.md" 2>/dev/null || {
        cat > "$CONFIG_DIR/README.md" << EOF
# CodeHunter v${VERSION}

Made with ❤️ by **Albert.C (@yz9yt)**

## 🔗 Links
- **GitHub**: [${GITHUB_REPO}](${GITHUB_REPO})
- **Twitter**: [@yz9yt](https://twitter.com/yz9yt)

## Quick Start

\`\`\`bash
# Basic scan
codehunter -f urls.txt -r secrets.txt -o results.json

# Pipe from tools
katana -u target.com | codehunter -f stdin -r api_endpoints.txt
\`\`\`

## Support the Project
- ⭐ Star on [GitHub](${GITHUB_REPO})
- 🐦 Follow [@yz9yt](https://twitter.com/yz9yt)
- 🔄 Share with the community

Made with ❤️ by Albert.C (@yz9yt)
EOF
    }
    
    echo -e "${GREEN}[INFO]${NC} Examples and documentation installed ✓"
}

test_installation() {
    echo -e "${YELLOW}[INFO]${NC} Testing installation..."
    
    if command -v codehunter &> /dev/null; then
        VERSION_OUTPUT=$(codehunter -version 2>/dev/null || echo "Version check failed")
        echo -e "${GREEN}[INFO]${NC} Installation test passed ✓"
        echo -e "${CYAN}[INFO]${NC} $VERSION_OUTPUT"
        
        # Test with example if available
        if [[ -f "$CONFIG_DIR/examples/urls.txt" && -f "$PATTERNS_DIR/secrets.txt" ]]; then
            echo -e "${CYAN}[INFO]${NC} Running quick test scan..."
            if codehunter -f "$CONFIG_DIR/examples/urls.txt" -r secrets.txt >/dev/null 2>&1; then
                echo -e "${GREEN}[INFO]${NC} Test scan completed successfully ✓"
            fi
        fi
    else
        echo -e "${RED}[ERROR]${NC} codehunter command not found in PATH!"
        echo "Installation may have failed"
        exit 1
    fi
}

show_completion() {
    echo -e "\n${BOLD}${GREEN}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║                INSTALLATION COMPLETE! 🎉                ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
    
    echo -e "${BOLD}${CYAN}CodeHunter v${VERSION} is ready! ⚡${NC}"
    echo -e "${PURPLE}Made with ❤️ by Albert.C (@yz9yt)${NC}"
    echo -e "${BLUE}GitHub: ${GITHUB_REPO}${NC}\n"
    
    echo -e "${YELLOW}🚀 Quick Start:${NC}"
    echo "  codehunter -f urls.txt -r secrets.txt -v"
    echo "  katana -u target.com | codehunter -f stdin -r api_endpoints.txt"
    echo "  codehunter -f wayback_urls.txt -r js_secrets.txt -js-only"
    echo ""
    
    echo -e "${YELLOW}📍 Locations:${NC}"
    echo "  • Binary: $INSTALL_DIR/codehunter"
    echo "  • Patterns: $PATTERNS_DIR"
    echo "  • Config: $CONFIG_DIR"
    echo "  • Examples: $CONFIG_DIR/examples/"
    echo ""
    
    echo -e "${YELLOW}🎯 Try it now:${NC}"
    echo "  codehunter -version"
    echo "  bash $CONFIG_DIR/examples/demo.sh"
    echo ""
    
    print_love_message
}

print_love_message() {
    echo -e "${BOLD}${PURPLE}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                      Thank you! 🙏                        ║"
    echo "║                                                           ║"
    echo "║  CodeHunter was crafted with love and late-night coffee  ║"
    echo "║              by Albert.C (@yz9yt)                        ║"
    echo "║                                                           ║"
    echo "║     If this tool helps you find bugs, please:            ║"
    echo "║       • ⭐ Star: ${GITHUB_REPO}         ║"
    echo "║       • 🐦 Follow: @yz9yt on Twitter                     ║"
    echo "║       • 🔄 Share with the Bug Bounty community           ║"
    echo "║                                                           ║"
    echo "║             Happy Bug Hunting! 🎯🔍                      ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

uninstall() {
    echo -e "${YELLOW}[INFO]${NC} Removing CodeHunter..."
    echo -e "${PURPLE}[INFO]${NC} We're sad to see you go! Made with ❤️ by ${AUTHOR}"
    
    # Remove binary
    $SUDO rm -f "$INSTALL_DIR/codehunter"
    
    # Remove system patterns
    $SUDO rm -rf "$PATTERNS_DIR"
    
    # Ask about user config
    read -p "Remove user configuration ($CONFIG_DIR)? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        echo -e "${GREEN}[INFO]${NC} User configuration removed"
    else
        echo -e "${YELLOW}[INFO]${NC} Keeping user configuration at $CONFIG_DIR"
    fi
    
    echo -e "${GREEN}[INFO]${NC} CodeHunter uninstalled"
    echo -e "${PURPLE}[INFO]${NC} Thanks for trying CodeHunter!"
    echo -e "${CYAN}[INFO]${NC} GitHub: ${GITHUB_REPO}"
    echo -e "${CYAN}[INFO]${NC} Follow @yz9yt for more security tools ✨"
}

print_help() {
    echo -e "${BOLD}CodeHunter Installer v${VERSION}${NC}"
    echo -e "Made with ❤️ by ${AUTHOR}"
    echo -e "GitHub: ${GITHUB_REPO}"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  install      Install CodeHunter (default)"
    echo "  uninstall    Remove CodeHunter"
    echo "  help         Show this help"
    echo ""
    echo "Installation Methods:"
    echo "  1. Prebuilt binary (fastest)"
    echo "  2. Build from source (requires Go)"
    echo ""
    echo "Examples:"
    echo "  curl -sSL ${GITHUB_RAW}/installer.sh | bash"
    echo "  bash installer.sh"
    echo "  bash installer.sh uninstall"
}

# Main installation function
install() {
    print_banner
    detect_os
    check_requirements
    get_latest_release
    
    # Try prebuilt binary first, fallback to source
    if download_prebuilt_binary; then
        echo -e "${GREEN}[INFO]${NC} Using prebuilt binary ⚡"
    else
        echo -e "${YELLOW}[INFO]${NC} Building from source..."
        build_from_source
    fi
    
    install_binary
    create_directories
    install_patterns
    install_examples
    test_installation
    show_completion
    
    # Cleanup
    cd /
    rm -rf "$TEMP_DIR" 2>/dev/null || true
}

# Parse command line arguments
case "${1:-install}" in
    "install"|"")
        install
        ;;
    "uninstall")
        uninstall
        ;;
    "help"|"-h"|"--help")
        print_help
        ;;
    *)
        echo -e "${RED}Error:${NC} Unknown command '$1'"
        echo -e "Run '$0 help' for usage information"
        echo -e "${PURPLE}Made with ❤️ by ${AUTHOR}${NC}"
        exit 1
        ;;
esac
