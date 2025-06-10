#!/bin/bash

# Telegram Daemon Installation Script
# This script installs the Telegram daemon as a systemd service

set -e

# Configuration
DAEMON_USER="telegram-daemon"
DAEMON_GROUP="telegram-daemon"
INSTALL_DIR="/opt/telegram-daemon"
SERVICE_FILE="/etc/systemd/system/telegram-daemon.service"

# Colors for output
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
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

log_info "Installing Telegram Daemon..."

# Create user and group
if ! id "$DAEMON_USER" &>/dev/null; then
    log_info "Creating user $DAEMON_USER..."
    useradd --system --shell /bin/false --home-dir "$INSTALL_DIR" --create-home "$DAEMON_USER"
else
    log_info "User $DAEMON_USER already exists"
fi

# Create installation directory
log_info "Creating installation directory..."
mkdir -p "$INSTALL_DIR"

# Copy binary
if [[ -f "./telegram-daemon" ]]; then
    log_info "Copying binary to $INSTALL_DIR..."
    cp ./telegram-daemon "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/telegram-daemon"
else
    log_error "Binary ./telegram-daemon not found. Please build it first with 'go build'"
    exit 1
fi

# Copy configuration example
if [[ -f "./config.yaml.example" ]]; then
    log_info "Copying configuration example..."
    cp ./config.yaml.example "$INSTALL_DIR/"
fi

# Set ownership
log_info "Setting ownership..."
chown -R "$DAEMON_USER:$DAEMON_GROUP" "$INSTALL_DIR"

# Install systemd service
log_info "Installing systemd service..."
cp ./telegram-daemon.service "$SERVICE_FILE"

# Reload systemd
log_info "Reloading systemd..."
systemctl daemon-reload

log_info "Installation complete!"
echo
log_warn "Next steps:"
echo "1. Copy and configure: cp $INSTALL_DIR/config.yaml.example $INSTALL_DIR/config.yaml"
echo "2. Edit the configuration: nano $INSTALL_DIR/config.yaml"
echo "3. Start the service: systemctl start telegram-daemon"
echo "4. Enable auto-start: systemctl enable telegram-daemon"
echo "5. Check status: systemctl status telegram-daemon"
echo "6. View logs: journalctl -u telegram-daemon -f"

