# Webhook Forge

Webhook Forge is a lightweight server for receiving webhook requests and creating flag files upon successful processing.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
  - [Building from Source](#building-from-source)
  - [Using Makefile](#using-makefile)
- [Configuration](#configuration)
  - [Main Configuration](#main-configuration)
  - [Logging Configuration](#logging-configuration)
  - [Admin Token Generation](#admin-token-generation)
  - [Reverse Proxy Configuration](#reverse-proxy-configuration)
  - [Enhanced Security with IP Restrictions](#enhanced-security-with-ip-restrictions)
- [API Endpoints](#api-endpoints)
  - [Webhook Management](#webhook-management)
  - [Webhook Invocation](#webhook-invocation)
  - [Admin Token Authentication](#admin-token-authentication)
  - [API Response Format](#api-response-format)
- [Usage Examples](#usage-examples)
  - [List All Webhooks](#list-all-webhooks)
  - [Create a Webhook](#create-a-webhook)
  - [Invoke a Webhook](#invoke-a-webhook)
- [Deployment](#deployment)
  - [Linux Service Setup](#linux-service-setup)
  - [Using Service Scripts](#using-service-scripts)
  - [Docker Deployment](#docker-deployment)
  - [Docker Hub](#docker-hub)
- [Project Structure](#project-structure)
- [License](#license)

## Features

- Simple API for creating and managing webhooks
- Secure token verification for authorizing calls
- Creation of flag files upon successful webhook invocation
- Easy configuration via JSON file
- Structured logging with file output and rotation
- Support for deployment behind a reverse proxy in a subdirectory

## Installation

### Building from Source

```bash
git clone git@github.com:mvandrew/webhook-forge.git
cd webhook-forge
go build -o webhook-forge ./cmd/server
```

### Using Makefile

The project includes a Makefile with various useful commands for building and running the application:

```bash
# Build both server and admin token generator
make build

# Build only the server
make build-server

# Build only the admin token generator
make build-admin-token

# Run the server
make run-server

# Generate admin token
make token
```

## Configuration

### Main Configuration

The configuration file is automatically created on first run in the `config/config.json` directory. You can modify the following parameters:

```json
{
  "server": {
    "host": "127.0.0.1",
    "port": 8080,
    "base_path": "",
    "admin_token": "admin-token"
  },
  "hooks": {
    "storage_path": "data/hooks.json",
    "flags_dir": "data/flags"
  },
  "log": {
    "level": "info",
    "format": "json",
    "file_path": "logs/webhook-forge.log",
    "max_size": 100,
    "max_backups": 5
  }
}
```

To set up the application, create your own configuration file based on the example:

```bash
cp config/config.example.json config/config.json
```

Then edit the `config/config.json` file according to your requirements. The example configuration file is tracked in git, while your local configuration file is ignored.

### Logging Configuration

The logging system supports output to files with automatic rotation to prevent excessive disk usage:

- `level`: Sets the minimum log level ("debug", "info", "warn", "error", "fatal")
- `format`: Log format ("json" or "text")
- `file_path`: Path to the log file (if omitted, logs go to stdout)
- `max_size`: Maximum size of each log file in megabytes before rotation
- `max_backups`: Maximum number of rotated log files to keep

When log files reach the maximum size, they are automatically rotated, and old files are named with a numeric suffix (e.g., `webhook-forge.log.1`, `webhook-forge.log.2`). Once the number of backups exceeds `max_backups`, the oldest files are removed.

### Admin Token Generation

For security reasons, you should generate a random admin token instead of using the default value. Use the provided admin token generator tool:

```bash
make token
```

This will generate a secure random token and ask for confirmation before saving it to your configuration file. If you already have a token in your configuration, you'll be shown both the current and new tokens before being asked to confirm the replacement.

### Reverse Proxy Configuration

If you're running the webhook-forge server behind a reverse proxy (like Nginx) at a subdirectory, you can use the `base_path` setting:

```json
{
  "server": {
    "host": "127.0.0.1",
    "port": 8080,
    "base_path": "/hooks"
  }
}
```

This will make all routes available at `/hooks/*` (e.g., `/hooks/api/hooks`, `/hooks/webhook/my-hook`), which can then be properly proxied.

#### Example Nginx Configuration

```nginx
server {
    listen 80;
    server_name example.com;

    location /hooks/ {
        proxy_pass http://127.0.0.1:8080/hooks/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Enhanced Security with IP Restrictions

For enhanced security, you can restrict access to webhook management endpoints based on IP addresses while keeping webhook invocation endpoints accessible from anywhere. This is particularly useful for production environments.

#### Example Nginx Configuration with IP Restrictions

```nginx
server {
    listen 80;
    server_name example.com;

    # Allow webhook invocation from anywhere
    location ~ ^/hooks/webhook/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Restrict webhook management API to specific IPs
    location ~ ^/hooks/api/ {
        # Allow only specific IP addresses
        allow 192.168.1.100;      # Admin workstation
        allow 10.0.0.0/24;        # Internal network
        deny all;                 # Deny everyone else
        
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## API Endpoints

### Webhook Management

- `GET /api/hooks` - List all webhooks (requires admin token)
- `GET /api/hooks/{id}` - Get information about a specific webhook (requires admin token)
- `POST /api/hooks` - Create a new webhook (requires admin token)
- `PUT /api/hooks/{id}` - Update an existing webhook (requires admin token)
- `DELETE /api/hooks/{id}` - Delete a webhook (requires admin token)

Note: If you've configured `base_path`, prepend it to these endpoints (e.g., `/hooks/api/hooks`).

### Webhook Invocation

- `POST /webhook/{id}?token=your-secret-token` - Trigger a webhook, creating the configured flag file

### Admin Token Authentication

All API endpoints (`GET`, `POST`, `PUT`, `DELETE`) require the `Authorization: Bearer <token>` header for authentication. The token must match the value defined in the server configuration.

### API Response Format

All API responses follow a consistent format:

```json
{
  "success": true|false,
  "data": {/* returned data object */},
  "errors": ["error message 1", "error message 2"]
}
```

When an operation is successful, the `success` field is `true` and the `data` field contains the result. When an error occurs, `success` is `false` and the `errors` field contains error messages.

## Usage Examples

### List All Webhooks

```bash
curl -X GET http://localhost:8080/api/hooks \
  -H "Authorization: Bearer admin-token"
```

### Create a Webhook

```bash
curl -X POST http://localhost:8080/api/hooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin-token" \
  -d '{
    "id": "my-webhook",
    "name": "My Webhook",
    "description": "Webhook for my project",
    "token": "your-secret-token",
    "flag_file": "my-project/flag.txt",
    "enabled": true
  }'
```

Note: The `token` field is optional. If not provided, a secure token will be automatically generated.

### Invoke a Webhook

```bash
curl -X POST "http://localhost:8080/webhook/my-webhook?token=your-secret-token"
```

After a successful invocation, the file will be created in the `data/flags/my-project/flag.txt` directory.

## Deployment

### Linux Service Setup

You can run webhook-forge as a system service on Linux using systemd. Create a service file at `/etc/systemd/system/webhook-forge.service`:

```ini
[Unit]
Description=Webhook Forge Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=youruser
WorkingDirectory=/path/to/webhook-forge
ExecStart=/path/to/webhook-forge/bin/webhook-forge

[Install]
WantedBy=multi-user.target
```

Replace `youruser` with the appropriate user and adjust the paths accordingly.

Enable and start the service:

```bash
sudo systemctl enable webhook-forge
sudo systemctl start webhook-forge
```

Check service status:

```bash
sudo systemctl status webhook-forge
```

View logs:

```bash
sudo journalctl -u webhook-forge
```

### Using Service Scripts

The project includes several scripts in the `scripts/` directory to simplify service installation and management.

#### Standard Service Installation

To install Webhook Forge as a systemd service:

```bash
sudo ./scripts/install-service.sh
```

This script supports the following command-line options:
- `-h, --help`: Show help message
- `-n, --name NAME`: Name of the service (default: "webhook-forge")
- `-u, --user USER`: User to run the service as (default: current user)
- `-d, --dir DIRECTORY`: Installation directory (default: current directory)
- `-c, --config PATH`: Path to config file (default: "INSTALL_DIR/config/config.json")
- `-e, --executable PATH`: Path to executable (default: "INSTALL_DIR/bin/server")

Example with custom settings:

```bash
sudo ./scripts/install-service.sh --name my-webhook --user webuser --executable /opt/webhook-forge/bin/webhook-forge
```

To uninstall the service:

```bash
sudo ./scripts/uninstall-service.sh
```

For custom service name, use the `--name` option:

```bash
sudo ./scripts/uninstall-service.sh --name my-webhook
```

#### Docker Service Installation

To install Webhook Forge as a Docker service managed by systemd:

```bash
sudo ./scripts/install-docker-service.sh
```

This script supports the following command-line options:
- `-h, --help`: Show help message
- `-n, --name NAME`: Name of the service (default: "webhook-forge-docker")
- `-d, --dir DIRECTORY`: Installation directory (default: current directory) 
- `-f, --file FILE`: Path to docker-compose file (default: "INSTALL_DIR/docker-compose.prod.yml")
- `-u, --user USER`: User to run the service as (default: current user)

Example with custom settings:

```bash
sudo ./scripts/install-docker-service.sh --name my-webhook-docker --file /opt/webhook-forge/docker-compose.yml
```

To uninstall the Docker service:

```bash
sudo ./scripts/uninstall-docker-service.sh
```

For custom service name, use the `--name` option:

```bash
sudo ./scripts/uninstall-docker-service.sh --name my-webhook-docker
```

### Docker Deployment

#### Using Docker

Build the Docker image:

```bash
make docker-build
```

Run the container:

```bash
make docker-run
```

Or for local-only access:

```bash
make docker-run-local
```

Stop the container:

```bash
make docker-stop
```

#### Using Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3'
services:
  webhook-forge:
    image: msav/webhook-forge:latest
    container_name: webhook-forge
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config
      - ./data:/app/data
      - ./logs:/app/logs
```

Start the service:

```bash
docker-compose up -d
```

For production deployment with the provided production configuration:

```bash
make docker-compose-prod-up
```

To stop the service:

```bash
make docker-compose-prod-down
```

#### Docker Hub

The webhook-forge Docker image is available on Docker Hub:
[https://hub.docker.com/r/msav/webhook-forge](https://hub.docker.com/r/msav/webhook-forge)

Pull the latest image:

```bash
docker pull msav/webhook-forge:latest
```

## Project Structure

The project is organized according to clean architecture principles:

- `cmd/server` - Application entry point
- `cmd/admin_token_generator` - Utility for generating admin tokens
- `internal/api` - HTTP handlers
- `internal/config` - Application configuration
- `internal/domain` - Data models and interfaces
- `internal/service` - Business logic
- `internal/storage` - Data storage
- `pkg/logger` - Logging
- `pkg/validator` - Data validation
- `scripts` - Service installation and management scripts

## License

GNU General Public License v3.0
