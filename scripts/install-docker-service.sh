#!/bin/bash

set -e

# Install webhook-forge docker service as a systemd service
# This script must be run with root privileges

# Default settings (can be overridden by environment variables)
SERVICE_NAME=${SERVICE_NAME:-"webhook-forge-docker"}
INSTALL_DIR=${INSTALL_DIR:-$(pwd)}
DOCKER_COMPOSE_FILE=${DOCKER_COMPOSE_FILE:-"$INSTALL_DIR/docker-compose.prod.yml"}
USER=${USER:-$(id -un)}

# Check if the script is running with sudo
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root or with sudo"
  exit 1
fi

echo "Installing webhook-forge docker service as a systemd service..."
echo "Service name: $SERVICE_NAME"
echo "Installation directory: $INSTALL_DIR"
echo "Docker compose file: $DOCKER_COMPOSE_FILE"
echo "User: $USER"

# Check if docker-compose file exists
if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
  echo "Error: Docker compose file not found at $DOCKER_COMPOSE_FILE"
  exit 1
fi

# Check if docker is installed
if ! command -v docker &> /dev/null; then
  echo "Error: Docker is not installed"
  exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
  echo "Error: Docker Compose is not installed"
  exit 1
fi

# Create systemd service file
cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=webhook-forge Docker service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$INSTALL_DIR
ExecStart=/usr/bin/docker-compose -f $DOCKER_COMPOSE_FILE up -d
ExecStop=/usr/bin/docker-compose -f $DOCKER_COMPOSE_FILE down
User=$USER

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Service installed successfully as $SERVICE_NAME.service"
echo "To start the service, run: sudo systemctl start $SERVICE_NAME"
echo "To enable it at boot, run: sudo systemctl enable $SERVICE_NAME"
echo "To check the service status: sudo systemctl status $SERVICE_NAME" 