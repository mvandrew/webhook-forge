#!/bin/bash

set -e

# Install webhook-forge as a systemd service
# This script must be run with root privileges

# Default settings (can be overridden by environment variables)
SERVICE_NAME=${SERVICE_NAME:-"webhook-forge"}
SERVICE_USER=${SERVICE_USER:-$(id -un)}
INSTALL_DIR=${INSTALL_DIR:-$(pwd)}
CONFIG_PATH=${CONFIG_PATH:-"$INSTALL_DIR/config/config.json"}
EXECUTABLE=${EXECUTABLE:-"$INSTALL_DIR/bin/server"}

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
  echo "Please build the application first or specify the correct path with EXECUTABLE environment variable"
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