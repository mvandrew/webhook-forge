#!/bin/bash

set -e

# Uninstall webhook-forge docker systemd service
# This script must be run with root privileges

# Default service name (can be overridden by environment variable)
SERVICE_NAME=${SERVICE_NAME:-"webhook-forge-docker"}

# Check if the script is running with sudo
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root or with sudo"
  exit 1
fi

echo "Uninstalling $SERVICE_NAME service..."

# Check if service exists
if ! systemctl list-unit-files | grep -q "$SERVICE_NAME.service"; then
  echo "Service $SERVICE_NAME does not exist"
  exit 0
fi

# Stop service if running
if systemctl is-active --quiet "$SERVICE_NAME"; then
  echo "Stopping $SERVICE_NAME service..."
  systemctl stop "$SERVICE_NAME"
fi

# Disable service
if systemctl is-enabled --quiet "$SERVICE_NAME"; then
  echo "Disabling $SERVICE_NAME service..."
  systemctl disable "$SERVICE_NAME"
fi

# Remove service file
echo "Removing service file..."
rm -f "/etc/systemd/system/$SERVICE_NAME.service"

# Reload systemd
systemctl daemon-reload
systemctl reset-failed

echo "$SERVICE_NAME service has been successfully uninstalled"
echo "Note: Docker containers may still be running. To stop them, run 'docker-compose down' in your project directory." 