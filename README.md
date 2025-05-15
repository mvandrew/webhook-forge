# Webhook Forge

Webhook Forge is a lightweight server for receiving webhook requests and creating flag files upon successful processing.

## Features

- Simple API for creating and managing webhooks
- Secure token verification for authorizing calls
- Creation of flag files upon successful webhook invocation
- Easy configuration via JSON file
- Structured logging

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
    "port": 8080
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

#### Example of Creating a Webhook

```bash
curl -X POST http://localhost:8080/api/hooks \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-webhook",
    "name": "My Webhook",
    "description": "Webhook for my project",
    "token": "your-secret-token",
    "flag_file": "my-project/flag.txt",
    "enabled": true
  }'
```

#### Invoking a Webhook

```bash
curl -X POST "http://localhost:8080/webhook/my-webhook?token=your-secret-token"
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
