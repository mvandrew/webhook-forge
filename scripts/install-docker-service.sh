#!/bin/bash

set -e

# Install webhook-forge docker service as a systemd service
# This script must be run with root privileges

# Default settings
SERVICE_NAME="webhook-forge-docker"
INSTALL_DIR=$(pwd)
DOCKER_COMPOSE_FILE=""
SERVICE_USER=$(id -un)

# Print usage information
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Install webhook-forge docker service as a systemd service"
  echo
  echo "Options:"
  echo "  -h, --help                 Show this help message"
  echo "  -n, --name NAME            Service name (default: webhook-forge-docker)"
  echo "  -d, --dir DIRECTORY        Installation directory (default: current directory)"
  echo "  -f, --file FILE            Path to docker-compose file (default: INSTALL_DIR/docker-compose.prod.yml)"
  echo "  -u, --user USER            User to run the service as (default: current user)"
  echo
  echo "Example:"
  echo "  $0 --name my-webhook-docker --file /opt/webhook-forge/docker-compose.yml"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      print_usage
      exit 0
      ;;
    -n|--name)
      SERVICE_NAME="$2"
      shift 2
      ;;
    -d|--dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -f|--file)
      DOCKER_COMPOSE_FILE="$2"
      shift 2
      ;;
    -u|--user)
      SERVICE_USER="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Set default for docker compose file if not provided
if [ -z "$DOCKER_COMPOSE_FILE" ]; then
  DOCKER_COMPOSE_FILE="$INSTALL_DIR/docker-compose.prod.yml"
fi

# Check if the script is running with sudo
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root or with sudo"
  exit 1
fi

echo "Installing webhook-forge docker service as a systemd service..."
echo "Service name: $SERVICE_NAME"
echo "Installation directory: $INSTALL_DIR"
echo "Docker compose file: $DOCKER_COMPOSE_FILE"
echo "User: $SERVICE_USER"

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
User=$SERVICE_USER

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Service installed successfully as $SERVICE_NAME.service"
echo "To start the service, run: sudo systemctl start $SERVICE_NAME"
echo "To enable it at boot, run: sudo systemctl enable $SERVICE_NAME"
echo "To check the service status: sudo systemctl status $SERVICE_NAME" 