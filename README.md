# Webhook Forge

Webhook Forge is a lightweight server for receiving webhook requests and creating flag files upon successful processing.

## Features

- Simple API for creating and managing webhooks
- Secure token verification for authorizing calls
- Creation of flag files upon successful webhook invocation
- Easy configuration via JSON file
- Structured logging
- Support for deployment behind a reverse proxy in a subdirectory

## Installation

```bash
git clone git@github.com:mvandrew/webhook-forge.git
cd webhook-forge
go build -o webhook-forge ./cmd/server
```

Alternatively, you can use the provided Makefile:

```bash
make build
```

## Configuration

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
    "format": "json"
  }
}
```

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

This configuration restricts access to the webhook management API (`/hooks/api/*`) to specific IP addresses, while allowing webhook invocation endpoints (`/hooks/webhook/*`) to be accessible from anywhere. This provides an additional layer of security by ensuring that only authorized systems can create, modify, or delete webhooks.

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

## Usage

### Starting the Server

```bash
./webhook-forge
```

Or using the Makefile:

```bash
make run-server
```

### Generating Admin Token

To generate a secure admin token for your webhook-forge server:

```bash
./bin/admin-token-generator
```

Or using the Makefile:

```bash
make token
```

The tool will generate a new random token and ask for your confirmation before saving it to the configuration file.

### API Endpoints

#### Webhook Management

- `GET /api/hooks` - get a list of all webhooks (requires admin token)
- `GET /api/hooks/{id}` - get information about a specific webhook (requires admin token)
- `POST /api/hooks` - create a new webhook (requires admin token)
- `PUT /api/hooks/{id}` - update an existing webhook (requires admin token)
- `DELETE /api/hooks/{id}` - delete a webhook (requires admin token)

Note: If you've configured `base_path`, prepend it to these endpoints (e.g., `/hooks/api/hooks`).

#### Example of Getting All Webhooks

```bash
curl -X GET http://localhost:8080/api/hooks \
  -H "Authorization: Bearer admin-token"
```

For a server with `base_path` set to `/hooks`:
```bash
curl -X GET http://localhost:8080/hooks/api/hooks \
  -H "Authorization: Bearer admin-token"
```

#### Example of Creating a Webhook

For a server running at the root:
```bash
curl -X POST http://localhost:8080/api/hooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin-token" \
  -d '{
    "id": "my-webhook",
    "name": "My Webhook",
    "description": "Webhook for my project",
    "token": "your-secret-token",  # Optional, will be generated automatically if not provided
    "flag_file": "my-project/flag.txt",
    "enabled": true
  }'
```

For a server with `base_path` set to `/hooks`:
```bash
curl -X POST http://localhost:8080/hooks/api/hooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin-token" \
  -d '{
    "id": "my-webhook",
    "name": "My Webhook",
    "description": "Webhook for my project",
    "token": "your-secret-token",  # Optional, will be generated automatically if not provided
    "flag_file": "my-project/flag.txt",
    "enabled": true
  }'
```

The response will contain the created webhook, including the generated token if one wasn't provided:

```json
{
  "success": true,
  "data": {
    "id": "my-webhook",
    "name": "My Webhook",
    "description": "Webhook for my project",
    "token": "your-secret-token",
    "flag_file": "my-project/flag.txt",
    "enabled": true,
    "created_at": "2023-04-18T12:34:56Z",
    "updated_at": "2023-04-18T12:34:56Z"
  }
}
```

#### Admin Token Authentication

All API endpoints (`GET`, `POST`, `PUT`, `DELETE`) require the `Authorization: Bearer <token>` header for authentication. The token must match the value defined in the server configuration.

If the token is missing or invalid, the API will return a `403 Forbidden` response:

```json
{
  "success": false,
  "errors": ["Admin authentication required"]
}
```

#### Automatic Token Generation

When creating a new webhook, if you don't specify a token, the system will automatically generate a secure token for you. This generated token will be returned in the response. Make sure to save it, as it will be needed to trigger the webhook.

#### Invoking a Webhook

For a server running at the root:
```bash
curl -X POST "http://localhost:8080/webhook/my-webhook?token=your-secret-token"
```

For a server with `base_path` set to `/hooks`:
```bash
curl -X POST "http://localhost:8080/hooks/webhook/my-webhook?token=your-secret-token"
```

After a successful invocation, the file will be created in the `data/flags/my-project/flag.txt` directory.

## Project Structure

The project is organized according to clean architecture principles:

- `cmd/server` - application entry point
- `cmd/admin_token_generator` - utility for generating admin tokens
- `internal/api` - HTTP handlers
- `internal/config` - application configuration
- `internal/domain` - data models and interfaces
- `internal/service` - business logic
- `internal/storage` - data storage
- `pkg/logger` - logging
- `pkg/validator` - data validation

## License

GNU General Public License v3.0

## Configuration Example

To set up the application, create your own configuration file based on the example:

```bash
cp config/config.example.json config/config.json
```

Then edit the `config/config.json` file according to your requirements. The example configuration file is tracked in git, while your local configuration file is ignored.
