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
git clone https://github.com/yourusername/webhook-forge.git
cd webhook-forge
go build -o webhook-forge ./cmd/server
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

### API Endpoints

#### Webhook Management

- `GET /api/hooks` - get a list of all webhooks
- `GET /api/hooks/{id}` - get information about a specific webhook
- `POST /api/hooks` - create a new webhook
- `PUT /api/hooks/{id}` - update an existing webhook
- `DELETE /api/hooks/{id}` - delete a webhook

Note: If you've configured `base_path`, prepend it to these endpoints (e.g., `/hooks/api/hooks`).

#### Example of Creating a Webhook

For a server running at the root:
```bash
curl -X POST http://localhost:8080/api/hooks \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: admin-token" \
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
  -H "X-Admin-Token: admin-token" \
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

All management endpoints (`POST`, `PUT`, `DELETE`) require the `X-Admin-Token` header for authentication. The token must match the value defined in the server configuration.

If the token is missing or invalid, the API will return a `403 Forbidden` response:

```json
{
  "success": false,
  "errors": ["Admin token required"]
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
