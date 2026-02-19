#!/bin/bash

# k0rdentd Installation Script
# This script downloads and installs the k0rdentd binary from GitHub releases

set -euo pipefail

# GitHub repository
GITHUB_REPO="belgaied2/k0rdentd"
INSTALL_PATH="/usr/local/bin/k0rdentd"

# Colors for output (only if terminal supports it)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root or with sudo"
        echo "Please run: sudo bash $0"
        exit 1
    fi
}

# Detect system architecture
detect_arch() {
    local arch
    arch=$(uname -m)

    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64)
            echo "arm64"
            ;;
        armv7l)
            echo "armv7"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            log_error "Supported architectures: x86_64 (amd64), aarch64 (arm64), armv7l (armv7)"
            exit 1
            ;;
    esac
}

# Get the version to install
get_version() {
    if [[ -n "${K0RDENTD_VERSION:-}" ]]; then
        # Remove 'v' prefix if present for consistency
        echo "${K0RDENTD_VERSION#v}"
    else
        # Fetch latest version from GitHub API
        # log_info "Fetching latest version from GitHub..."
        local latest
        latest=$(curl -fsSL "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | \
                 grep '"tag_name":' | \
                 sed -E 's/.*"v?([^"]+)".*/\1/')

        if [[ -z "$latest" ]]; then
            log_error "Failed to determine latest version"
            exit 1
        fi

        echo "$latest"
    fi
}

# Download and install the binary
download_and_install() {
    local version="$1"
    local arch="$2"
    local filename="k0rdentd-v${version}-linux-${arch}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/v${version}/${filename}"
    local tmp_dir
    tmp_dir=$(mktemp -d)

    # Cleanup function
    cleanup() {
        if [[ -d "$tmp_dir" ]]; then
            rm -rf "$tmp_dir"
        fi
    }
    trap cleanup EXIT

    log_info "Downloading k0rdentd v${version} for ${arch}..."
    log_info "URL: $download_url"

    # Download to temp directory
    if ! curl -fsSL "$download_url" -o "${tmp_dir}/${filename}"; then
        log_error "Failed to download $filename"
        log_error "Please check the version and architecture are correct"
        exit 1
    fi

    log_info "Download complete. Extracting..."

    # Extract the archive
    if ! tar -xzf "${tmp_dir}/${filename}" -C "$tmp_dir"; then
        log_error "Failed to extract archive"
        exit 1
    fi

    # Find the binary (it might be in a subdirectory)
    local binary_path
    binary_path=$(find "$tmp_dir" -name "k0rdentd" -type f | head -1)

    if [[ -z "$binary_path" ]]; then
        log_error "Could not find k0rdentd binary in the extracted archive"
        exit 1
    fi

    log_info "Installing to ${INSTALL_PATH}..."

    # Move to installation directory
    if ! mv "$binary_path" "$INSTALL_PATH"; then
        log_error "Failed to install binary to $INSTALL_PATH"
        exit 1
    fi

    # Set executable permissions
    chmod +x "$INSTALL_PATH"

    # Cleanup is handled by trap

    log_info "k0rdentd v${version} installed successfully!"
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."

    if ! command -v k0rdentd &> /dev/null; then
        log_warn "k0rdentd is not in your PATH"
        log_warn "You may need to start a new shell or run: export PATH=$PATH:/usr/local/bin"
        return
    fi

    local installed_version
    installed_version=$(k0rdentd --version 2>/dev/null || echo "unknown")
    log_info "Installed: $installed_version"
}

# Print next steps
print_next_steps() {
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "To deploy k0s and k0rdent on your VM, run:"
    echo ""
    echo "    sudo k0rdentd install"
    echo ""
    echo "For more options, see:"
    echo "    k0rdentd --help"
    echo ""
}

# Main function
main() {
    echo "=========================================="
    echo "  k0rdentd Installer"
    echo "=========================================="
    echo ""

    # Check privileges
    check_root

    # Detect architecture
    ARCH=$(detect_arch)
    log_info "Detected architecture: $ARCH"

    # Get version
    VERSION=$(get_version)
    log_info "Installing version: $VERSION"

    # Download and install
    download_and_install "$VERSION" "$ARCH"

    # Verify
    verify_installation

    # Print next steps
    print_next_steps
}

# Run main function
main "$@"
