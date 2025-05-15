#!/bin/bash

set -e

# Install webhook-forge as a systemd service
# This script must be run with root privileges

# Default settings
SERVICE_NAME="webhook-forge"
SERVICE_USER=$(id -un)
INSTALL_DIR=$(pwd)
CONFIG_PATH=""
EXECUTABLE=""

# Print usage information
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Install webhook-forge as a systemd service"
  echo
  echo "Options:"
  echo "  -h, --help               Show this help message"
  echo "  -n, --name NAME          Service name (default: webhook-forge)"
  echo "  -u, --user USER          Service user (default: current user)"
  echo "  -d, --dir DIRECTORY      Installation directory (default: current directory)"
  echo "  -c, --config PATH        Path to config file (default: INSTALL_DIR/config/config.json)"
  echo "  -e, --executable PATH    Path to executable (default: INSTALL_DIR/bin/server)"
  echo
  echo "Example:"
  echo "  $0 --name my-webhook --user webuser --executable /opt/webhook-forge/bin/server"
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
    -u|--user)
      SERVICE_USER="$2"
      shift 2
      ;;
    -d|--dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -c|--config)
      CONFIG_PATH="$2"
      shift 2
      ;;
    -e|--executable)
      EXECUTABLE="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Set defaults for paths if not provided
if [ -z "$CONFIG_PATH" ]; then
  CONFIG_PATH="$INSTALL_DIR/config/config.json"
fi

if [ -z "$EXECUTABLE" ]; then
  EXECUTABLE="$INSTALL_DIR/bin/server"
fi

# Check if the script is running with sudo
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root or with sudo"
  exit 1
fi

echo "Installing webhook-forge as a systemd service..."
echo "Service name: $SERVICE_NAME"
echo "Service user: $SERVICE_USER"
echo "Installation directory: $INSTALL_DIR"
echo "Config path: $CONFIG_PATH"
echo "Executable: $EXECUTABLE"

# Check if executable exists
if [ ! -f "$EXECUTABLE" ]; then
  echo "Error: Executable not found at $EXECUTABLE"
  echo "Please build the application first or specify the correct path with --executable option"
  exit 1
fi

# Check if config file exists
if [ ! -f "$CONFIG_PATH" ]; then
  echo "Warning: Config file not found at $CONFIG_PATH"
  echo "Make sure to create it before starting the service"
fi

# Create systemd service file
cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=webhook-forge service
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$EXECUTABLE
Restart=on-failure
RestartSec=10
Environment=CONFIG_PATH=$CONFIG_PATH

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Service installed successfully as $SERVICE_NAME.service"
echo "To start the service, run: sudo systemctl start $SERVICE_NAME"
echo "To enable it at boot, run: sudo systemctl enable $SERVICE_NAME"
echo "To check the service status: sudo systemctl status $SERVICE_NAME" 