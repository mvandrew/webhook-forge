#!/bin/bash

set -e

# Uninstall webhook-forge systemd service
# This script must be run with root privileges

# Default service name
SERVICE_NAME="webhook-forge"

# Print usage information
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Uninstall webhook-forge systemd service"
  echo
  echo "Options:"
  echo "  -h, --help           Show this help message"
  echo "  -n, --name NAME      Service name to uninstall (default: webhook-forge)"
  echo
  echo "Example:"
  echo "  $0 --name my-webhook"
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
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

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