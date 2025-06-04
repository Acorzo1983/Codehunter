#!/usr/bin/env bash
# CodeHunter v2.5.1 - Linux Native Installer
# Made with ❤️ by Albert.C @yz9yt
# https://github.com/Acorzo1983/Codehunter

# --- Configuration ---
VERSION="2.5.1" # Asegúrate que coincida con la versión que se va a instalar
APP_NAME="codehunter"
DEFAULT_INSTALL_DIR="/usr/local/bin"
DEFAULT_PATTERNS_DIR="/usr/share/${APP_NAME}"
USER_INSTALL_DIR="$HOME/.local/bin"
USER_PATTERNS_DIR="$HOME/.local/share/${APP_NAME}"

# --- Colors and Formatting ---
R='\033[0;31m' # Red
G='\033[0;32m' # Green
Y='\033[1;33m' # Yellow
B='\033[0;34m' # Blue
P='\033[0;35m' # Purple
C='\033[0;36m' # Cyan
BOLD='\033[1m'
NC='\033[0m'   # No Color

# --- Helper Functions ---
info() {
    echo -e "${C}[INFO]${NC} $1"
}

warn() {
    echo -e "${Y}[WARN]${NC} $1" >&2
}

error() {
    echo -e "${R}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${G}[SUCCESS]${NC} $1"
}

# Función para salir en caso de error
# Uso: die "Mensaje de error" [código_de_salida_opcional]
die() {
    error "$1"
    cleanup # Limpia antes de salir
    exit "${2:-1}" # Sale con el código de salida proporcionado o 1 por defecto
}

# --- Script Setup ---
# Salir inmediatamente si un comando falla
set -o errexit
# Tratar variables no establecidas como un error
set -o nounset
# Las tuberías fallan si cualquier comando en ellas falla, no solo el último
set -o pipefail

# Directorios de instalación actuales (se pueden modificar con --local)
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
PATTERNS_DIR="$DEFAULT_PATTERNS_DIR"
SUDO_CMD="" # Comando sudo a usar, vacío si no se necesita

# --- Main Functions ---
banner() {
    clear
    echo -e "${BOLD}${P}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║              CodeHunter v${VERSION} Installer                 ║"
    echo "║               Ultra-Fast Bug Bounty Scanner                 ║"
    echo "║                  🏴‍☠️ Kali Linux Ready 🏴‍☠️                  ║"
    echo "║                                                              ║"
    echo -e "║              Made with ${R}❤️${P} by Albert.C @yz9yt              ║"
    echo "║           github.com/Acorzo1983/Codehunter                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

check_os() {
    info "Checking system..."
    if [[ "$(uname -s)" != "Linux" && "$(uname -s)" != "Darwin" ]]; then
        die "Unsupported OS: $(uname -s). Only Linux and macOS are supported."
    fi
    success "Operating system: $(uname -s)"
    if [[ "$(uname -s)" == "Linux" && -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        source /etc/os-release
        info "Distribution: $PRETTY_NAME"
    fi
}

check_deps() {
    info "Checking dependencies..."
    local dep_missing=0

    if ! command -v go &>/dev/null; then
        error "Go (golang) not found."
        warn "Please install Go (e.g., sudo apt install golang-go) and try again."
        dep_missing=1
    else
        success "Go $(go version | awk '{print $3}') found."
    fi

    if ! command -v git &>/dev/null; then
        error "Git not found."
        warn "Please install Git (e.g., sudo apt install git) and try again."
        dep_missing=1
    else
        success "Git found."
    fi

    if [[ "$dep_missing" -eq 1 ]]; then
        die "One or more dependencies are missing. Please install them and re-run the installer."
    fi
}

verify_project_files() {
    info "Verifying project structure..."
    local essential_files=("main.go" "go.mod" "patterns/secrets.txt") #
    for file in "${essential_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            die "Missing essential project file: $file. Ensure you are in the project's root directory."
        fi
    done
    if [[ ! -d "patterns" ]]; then
        die "Missing 'patterns' directory. Ensure you are in the project's root directory."
    fi
    success "Project files verified."
}

determine_sudo() {
    if [[ "$INSTALL_DIR" == "$DEFAULT_INSTALL_DIR" || "$PATTERNS_DIR" == "$DEFAULT_PATTERNS_DIR" ]]; then
        if [[ "$(id -u)" -ne 0 ]]; then # Si no es root
            if command -v sudo &>/dev/null; then
                info "Installation to system directories requires root privileges."
                SUDO_CMD="sudo"
                # Probar sudo para pedir la contraseña al principio si es necesario
                if ! $SUDO_CMD -v; then
                    die "Failed to acquire sudo privileges. Please run as root or use --local."
                fi
            else
                die "sudo command not found, but needed for system-wide installation. Please run as root or use --local."
            fi
        fi
    fi
}

build_app() {
    info "Building ${APP_NAME}..."
    # Limpiar binario antiguo si existe
    if [[ -f "${APP_NAME}" ]]; then
        info "Removing existing '${APP_NAME}' binary before build."
        rm -f "${APP_NAME}"
    fi

    # Construir la aplicación Go
    # Usar CGO_ENABLED=0 para un binario estático si no hay dependencias C
    # GOOS y GOARCH se pueden establecer aquí para cross-compilation si es necesario
    if CGO_ENABLED=0 go build -ldflags="-s -w" -o "${APP_NAME}" main.go; then #
        success "${APP_NAME} binary built successfully."
    else
        die "Failed to build ${APP_NAME}."
    fi
    chmod +x "${APP_NAME}" #
}

install_app() {
    info "Installing ${APP_NAME} to ${INSTALL_DIR}..."
    if [[ ! -d "$INSTALL_DIR" ]]; then
        info "Creating installation directory: ${INSTALL_DIR}"
        # shellcheck disable=SC2086 # $SUDO_CMD puede estar vacío
        $SUDO_CMD mkdir -p "$INSTALL_DIR" || die "Failed to create directory: ${INSTALL_DIR}"
    fi
    # shellcheck disable=SC2086
    $SUDO_CMD cp "${APP_NAME}" "${INSTALL_DIR}/" || die "Failed to copy binary to ${INSTALL_DIR}/"
    success "${APP_NAME} binary installed."

    info "Installing patterns to ${PATTERNS_DIR}/patterns ..."
    if [[ ! -d "${PATTERNS_DIR}/patterns" ]]; then
        info "Creating patterns directory: ${PATTERNS_DIR}/patterns"
        # shellcheck disable=SC2086
        $SUDO_CMD mkdir -p "${PATTERNS_DIR}/patterns" || die "Failed to create directory: ${PATTERNS_DIR}/patterns"
    fi
    # shellcheck disable=SC2086
    $SUDO_CMD cp -r patterns/* "${PATTERNS_DIR}/patterns/" || die "Failed to copy patterns to ${PATTERNS_DIR}/patterns/"
    success "Patterns installed."
}

verify_installation() {
    info "Verifying installation..."
    if ! command -v "${APP_NAME}" &>/dev/null; then
        warn "${APP_NAME} not found in PATH immediately after installation."
        warn "You might need to open a new terminal session or add ${INSTALL_DIR} to your PATH."
        warn "Example for bash: echo 'export PATH=\$PATH:${INSTALL_DIR}' >> ~/.bashrc && source ~/.bashrc"
        # No fallar aquí, ya que el PATH puede no actualizarse instantáneamente.
    else
        success "${APP_NAME} found in PATH: $(command -v "${APP_NAME}")"
    fi

    if [[ ! -f "${PATTERNS_DIR}/patterns/secrets.txt" ]]; then
        die "Patterns verification failed. ${PATTERNS_DIR}/patterns/secrets.txt not found."
    else
        success "Patterns directory seems to be correctly populated."
    fi
}

uninstall_app() {
    banner
    info "Starting CodeHunter Uninstallation..."

    # Determinar si se usó --local para la instalación a desinstalar
    # Esto es una simplificación; un desinstalador robusto necesitaría saber la ruta exacta
    # o permitir al usuario especificarla.
    local uninstall_paths_to_check=(
        "${DEFAULT_INSTALL_DIR}/${APP_NAME}"
        "${USER_INSTALL_DIR}/${APP_NAME}"
    )
    local app_binary_found=""
    for path_to_check in "${uninstall_paths_to_check[@]}"; do
        if [[ -f "$path_to_check" ]]; then
            app_binary_found="$path_to_check"
            break
        fi
    done

    if [[ -z "$app_binary_found" ]]; then
        warn "${APP_NAME} does not seem to be installed in standard locations. No action taken."
        exit 0
    fi

    # Determinar SUDO_CMD para la desinstalación
    SUDO_CMD=""
    if [[ "$app_binary_found" == "${DEFAULT_INSTALL_DIR}/${APP_NAME}" && "$(id -u)" -ne 0 ]]; then
        if command -v sudo &>/dev/null; then
            SUDO_CMD="sudo"
            if ! $SUDO_CMD -v; then
                 die "Failed to acquire sudo privileges for uninstallation."
            fi
        else
            die "sudo not found, but needed to uninstall from ${DEFAULT_INSTALL_DIR}. Please run as root."
        fi
    fi
    
    # Confirmar con el usuario
    read -r -p "$(echo -e "${Y}[CONFIRM]${NC} Are you sure you want to uninstall ${APP_NAME} from ${app_binary_found%/*} and its patterns? (y/N): ")" confirmation
    if [[ "${confirmation,,}" != "y" ]]; then
        info "Uninstallation cancelled by user."
        exit 0
    fi

    # Establecer directorios de desinstalación basados en el binario encontrado
    if [[ "$app_binary_found" == "${DEFAULT_INSTALL_DIR}/${APP_NAME}" ]]; then
        UNINSTALL_BIN_PATH="${DEFAULT_INSTALL_DIR}/${APP_NAME}"
        UNINSTALL_PATTERNS_PATH="${DEFAULT_PATTERNS_DIR}"
    else
        UNINSTALL_BIN_PATH="${USER_INSTALL_DIR}/${APP_NAME}"
        UNINSTALL_PATTERNS_PATH="${USER_PATTERNS_DIR}"
    fi

    info "Removing ${APP_NAME} binary from ${UNINSTALL_BIN_PATH}..."
    if [[ -f "$UNINSTALL_BIN_PATH" ]]; then
        # shellcheck disable=SC2086
        $SUDO_CMD rm -f "$UNINSTALL_BIN_PATH" || warn "Could not remove binary: $UNINSTALL_BIN_PATH. Manual removal might be needed."
        success "Binary removed."
    else
        warn "Binary not found at ${UNINSTALL_BIN_PATH}. Skipping."
    fi

    info "Removing patterns directory from ${UNINSTALL_PATTERNS_PATH}..."
    if [[ -d "$UNINSTALL_PATTERNS_PATH" ]]; then
        # shellcheck disable=SC2086
        $SUDO_CMD rm -rf "$UNINSTALL_PATTERNS_PATH" || warn "Could not remove patterns directory: $UNINSTALL_PATTERNS_PATH. Manual removal might be needed."
        success "Patterns directory removed."
    else
        warn "Patterns directory not found at ${UNINSTALL_PATTERNS_PATH}. Skipping."
    fi

    success "CodeHunter uninstallation attempted. Please check for any remaining files if warnings occurred."
}


cleanup() {
    # Limpiar binario compilado si existe en el directorio actual
    if [[ -f "./${APP_NAME}" ]]; then
        info "Cleaning up temporary build file: ./${APP_NAME}"
        rm -f "./${APP_NAME}"
    fi
}

show_success_message() {
    echo -e "\n${BOLD}${G}🎉 INSTALLATION COMPLETE! 🎉${NC}\n"

    echo -e "${BOLD}${Y}📍 Installation Details:${NC}"
    echo -e "${G}Binary:${NC} ${INSTALL_DIR}/${APP_NAME}" #
    echo -e "${G}Patterns:${NC} ${PATTERNS_DIR}/patterns/" #

    echo -e "\n${BOLD}${C}🚀 Quick Start:${NC}"
    echo -e "${B}Basic scan:${NC}"
    echo "  ${APP_NAME} -r secrets.txt -l urls.txt -o results.txt" #

    echo -e "\n${B}Bug Bounty workflow:${NC}"
    echo "  subfinder -d tesla.com | httpx | ${APP_NAME} -r api_endpoints.txt" #
    echo "  katana -u tesla.com | ${APP_NAME} -r secrets.txt" #
    echo "  waybackurls target.com | ${APP_NAME} -r js_secrets.txt" #

    echo -e "\n${BOLD}${P}📋 Default Pattern Location (if system-wide):${NC}"
    echo "  ${DEFAULT_PATTERNS_DIR}/patterns/"
    echo "  Example: ${APP_NAME} -r ${DEFAULT_PATTERNS_DIR}/patterns/secrets.txt ..." #

    if [[ "$INSTALL_DIR" == "$USER_INSTALL_DIR" && ":$PATH:" != *":${USER_INSTALL_DIR}:"* ]]; then
        warn "Your PATH environment variable does not seem to include ${USER_INSTALL_DIR}."
        warn "Please add it by running: echo 'export PATH=\$PATH:${USER_INSTALL_DIR}' >> ~/.bashrc && source ~/.bashrc (or equivalent for your shell)"
    fi

    echo -e "\n${BOLD}${P}🏴‍☠️ Made with ❤️ by Albert.C @yz9yt 🏴‍☠️${NC}" #
    echo -e "${C}GitHub: https://github.com/Acorzo1983/Codehunter${NC}" #

    echo -e "\n${BOLD}${G}Happy Bug Hunting! 🎯${NC}" #
}

show_help() {
    banner
    echo -e "${BOLD}CodeHunter v${VERSION} Installer/Uninstaller${NC}"
    echo ""
    echo -e "${Y}Usage:${NC} $0 [options]"
    echo ""
    echo -e "${Y}Options:${NC}"
    echo "  -h, --help         Show this help message."
    echo "  -l, --local        Install to user's local directory (~/.local/bin and ~/.local/share)."
    echo "                     This does not require sudo."
    echo "      --uninstall    Attempt to uninstall CodeHunter from system or local directories."
    echo ""
    echo -e "${C}Default system-wide installation paths:${NC}"
    echo "  Binary:   ${DEFAULT_INSTALL_DIR}"
    echo "  Patterns: ${DEFAULT_PATTERNS_DIR}"
    echo ""
    echo -e "${C}Default local installation paths (with --local):${NC}"
    echo "  Binary:   ${USER_INSTALL_DIR}"
    echo "  Patterns: ${USER_PATTERNS_DIR}"
    echo ""
    echo -e "${C}Made with ❤️ by Albert.C @yz9yt${NC}" #
}

# --- Script Execution ---
main() {
    # Capturar la señal de error y ejecutar 'die'
    trap 'die "An unexpected error occurred. Line: $LINENO"' ERR

    # Parse command line arguments
    LOCAL_INSTALL=false
    UNINSTALL_MODE=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -l|--local)
                LOCAL_INSTALL=true
                shift
                ;;
            --uninstall)
                UNINSTALL_MODE=true
                shift
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    if [[ "$UNINSTALL_MODE" == true ]]; then
        uninstall_app
        exit 0
    fi

    # Si es instalación local, cambiar las rutas
    if [[ "$LOCAL_INSTALL" == true ]]; then
        INSTALL_DIR="$USER_INSTALL_DIR"
        PATTERNS_DIR="$USER_PATTERNS_DIR"
        info "Local installation selected. Using user directories."
    fi

    banner
    check_os
    check_deps
    verify_project_files
    determine_sudo # Determinar si se necesita sudo antes de construir y copiar
    build_app
    install_app
    verify_installation
    cleanup # Limpiar después de una instalación exitosa
    show_success_message

    # Resetear el trap de error al final de una ejecución exitosa
    trap - ERR
}

# Ejecutar la función principal con todos los argumentos pasados al script
main "$@"
