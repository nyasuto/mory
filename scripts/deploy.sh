#!/bin/bash
set -euo pipefail

# Mory Server Deployment Script
# For Ubuntu/Debian home server deployment

# Configuration
DEPLOY_USER="mory"
DEPLOY_DIR="/opt/mory-server"
SERVICE_NAME="mory-server"
PYTHON_VERSION="3.11"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

check_dependencies() {
    log_info "Checking system dependencies..."
    
    # Check Python version
    if ! command -v python${PYTHON_VERSION} &> /dev/null; then
        log_error "Python ${PYTHON_VERSION} is required but not installed"
        log_info "Install with: sudo apt update && sudo apt install python${PYTHON_VERSION} python${PYTHON_VERSION}-venv"
        exit 1
    fi
    
    # Check for required packages
    local packages=("curl" "git" "sqlite3")
    for package in "${packages[@]}"; do
        if ! command -v "$package" &> /dev/null; then
            log_error "$package is required but not installed"
            log_info "Install with: sudo apt update && sudo apt install $package"
            exit 1
        fi
    done
    
    log_info "All dependencies satisfied"
}

create_user() {
    log_info "Creating service user..."
    
    if id "$DEPLOY_USER" &>/dev/null; then
        log_warn "User $DEPLOY_USER already exists"
    else
        useradd --system --create-home --shell /bin/bash "$DEPLOY_USER"
        log_info "Created user: $DEPLOY_USER"
    fi
}

setup_directories() {
    log_info "Setting up directories..."
    
    # Create deployment directory
    mkdir -p "$DEPLOY_DIR"
    mkdir -p "${DEPLOY_DIR}/data"
    mkdir -p "${DEPLOY_DIR}/logs"
    mkdir -p "${DEPLOY_DIR}/backups"
    
    # Set ownership
    chown -R "$DEPLOY_USER:$DEPLOY_USER" "$DEPLOY_DIR"
    
    # Set permissions
    chmod 755 "$DEPLOY_DIR"
    chmod 750 "${DEPLOY_DIR}/data"
    chmod 750 "${DEPLOY_DIR}/logs"
    chmod 750 "${DEPLOY_DIR}/backups"
    
    log_info "Directories created and configured"
}

install_uv() {
    log_info "Installing uv package manager..."
    
    # Install uv for the mory user
    sudo -u "$DEPLOY_USER" bash -c "curl -LsSf https://astral.sh/uv/install.sh | sh"
    
    # Add uv to PATH for the service user
    if ! grep -q "uv" "/home/$DEPLOY_USER/.bashrc"; then
        echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> "/home/$DEPLOY_USER/.bashrc"
    fi
    
    log_info "uv installed successfully"
}

deploy_application() {
    log_info "Deploying application code..."
    
    # Copy application files
    if [[ -f "pyproject.toml" ]]; then
        # We're in the project directory
        cp -r app/ "$DEPLOY_DIR/"
        cp pyproject.toml "$DEPLOY_DIR/"
        cp README.md "$DEPLOY_DIR/" 2>/dev/null || true
        cp -r scripts/ "$DEPLOY_DIR/" 2>/dev/null || true
    else
        log_error "Must run this script from the mory-server project root"
        exit 1
    fi
    
    # Set ownership
    chown -R "$DEPLOY_USER:$DEPLOY_USER" "$DEPLOY_DIR"
    
    log_info "Application code deployed"
}

setup_python_environment() {
    log_info "Setting up Python environment..."
    
    # Create and setup virtual environment as mory user
    sudo -u "$DEPLOY_USER" bash -c "
        cd '$DEPLOY_DIR'
        export PATH=\"/home/$DEPLOY_USER/.cargo/bin:\$PATH\"
        uv venv --python python${PYTHON_VERSION}
        uv sync
    "
    
    log_info "Python environment configured"
}

setup_systemd_service() {
    log_info "Setting up systemd service..."
    
    # Copy service file
    if [[ -f "scripts/mory-server.service" ]]; then
        cp "scripts/mory-server.service" "/etc/systemd/system/"
    else
        log_error "Service file not found: scripts/mory-server.service"
        exit 1
    fi
    
    # Reload systemd and enable service
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    
    log_info "Systemd service configured"
}

setup_environment() {
    log_info "Setting up environment configuration..."
    
    # Create environment file if it doesn't exist
    if [[ ! -f "${DEPLOY_DIR}/.env" ]]; then
        cat > "${DEPLOY_DIR}/.env" << EOF
# Mory Server Configuration
MORY_HOST=0.0.0.0
MORY_PORT=8080
MORY_DEBUG=false
MORY_DATA_DIR=${DEPLOY_DIR}/data

# Database
MORY_DATABASE_URL=sqlite:///${DEPLOY_DIR}/data/mory.db

# Semantic Search (optional - add your OpenAI API key)
# OPENAI_API_KEY=your_openai_api_key_here
MORY_SEMANTIC_SEARCH_ENABLED=true
MORY_OPENAI_MODEL=text-embedding-3-large
MORY_HYBRID_SEARCH_WEIGHT=0.7

# Obsidian Integration (optional)
# MORY_OBSIDIAN_VAULT_PATH=/path/to/your/obsidian/vault
EOF
        
        # Set secure permissions
        chown "$DEPLOY_USER:$DEPLOY_USER" "${DEPLOY_DIR}/.env"
        chmod 600 "${DEPLOY_DIR}/.env"
        
        log_warn "Environment file created at ${DEPLOY_DIR}/.env"
        log_warn "Please edit this file to add your OpenAI API key and other settings"
    else
        log_info "Environment file already exists"
    fi
}

setup_log_rotation() {
    log_info "Setting up log rotation..."
    
    cat > "/etc/logrotate.d/mory-server" << EOF
/opt/mory-server/logs/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0644 $DEPLOY_USER $DEPLOY_USER
    postrotate
        systemctl reload-or-restart $SERVICE_NAME
    endscript
}
EOF
    
    log_info "Log rotation configured"
}

create_backup_script() {
    log_info "Creating backup script..."
    
    cat > "${DEPLOY_DIR}/scripts/backup.sh" << 'EOF'
#!/bin/bash
# Mory Server Backup Script

set -euo pipefail

BACKUP_DIR="/opt/mory-server/backups"
DATA_DIR="/opt/mory-server/data"
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="mory_backup_${DATE}.tar.gz"

# Create backup
echo "Creating backup: $BACKUP_FILE"
tar -czf "${BACKUP_DIR}/${BACKUP_FILE}" -C "$DATA_DIR" .

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "mory_backup_*.tar.gz" -mtime +7 -delete

echo "Backup completed: ${BACKUP_DIR}/${BACKUP_FILE}"
EOF
    
    chmod +x "${DEPLOY_DIR}/scripts/backup.sh"
    chown "$DEPLOY_USER:$DEPLOY_USER" "${DEPLOY_DIR}/scripts/backup.sh"
    
    # Setup daily backup cron job
    (crontab -u "$DEPLOY_USER" -l 2>/dev/null || true; echo "0 2 * * * ${DEPLOY_DIR}/scripts/backup.sh") | crontab -u "$DEPLOY_USER" -
    
    log_info "Backup script and cron job created"
}

start_service() {
    log_info "Starting Mory server..."
    
    systemctl start "$SERVICE_NAME"
    sleep 3
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_info "Service started successfully"
        systemctl status "$SERVICE_NAME" --no-pager
    else
        log_error "Service failed to start"
        systemctl status "$SERVICE_NAME" --no-pager
        exit 1
    fi
}

print_summary() {
    log_info "Deployment completed successfully!"
    echo
    echo "📋 Deployment Summary:"
    echo "  Service User:      $DEPLOY_USER"
    echo "  Install Directory: $DEPLOY_DIR"
    echo "  Data Directory:    ${DEPLOY_DIR}/data"
    echo "  Config File:       ${DEPLOY_DIR}/.env"
    echo "  Service:           $SERVICE_NAME"
    echo
    echo "🔧 Useful Commands:"
    echo "  Service Status:    sudo systemctl status $SERVICE_NAME"
    echo "  View Logs:         sudo journalctl -u $SERVICE_NAME -f"
    echo "  Restart Service:   sudo systemctl restart $SERVICE_NAME"
    echo "  Manual Backup:     sudo -u $DEPLOY_USER ${DEPLOY_DIR}/scripts/backup.sh"
    echo
    echo "🌐 Access:"
    echo "  API:               http://localhost:8080"
    echo "  Documentation:     http://localhost:8080/docs"
    echo "  Health Check:      http://localhost:8080/api/health"
    echo
    echo "⚠️  Next Steps:"
    echo "  1. Edit ${DEPLOY_DIR}/.env to add your OpenAI API key"
    echo "  2. Configure firewall if needed: sudo ufw allow 8080"
    echo "  3. Test the service: curl http://localhost:8080/api/health"
}

main() {
    log_info "Starting Mory Server deployment..."
    
    check_root
    check_dependencies
    create_user
    setup_directories
    install_uv
    deploy_application
    setup_python_environment
    setup_systemd_service
    setup_environment
    setup_log_rotation
    create_backup_script
    start_service
    print_summary
}

# Run main function
main "$@"