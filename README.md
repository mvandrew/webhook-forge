# Webhook Forge

A minimalist, configurable webhook receiver service written in Go that creates flag files on your server when triggered.

## Overview

Webhook Forge is a lightweight, versatile webhook receiver designed to bridge webhook events with local file system actions. When properly configured webhooks are triggered, the service creates predefined flag files at specified locations on your server, which can then be monitored by other applications or scripts.

## Features

- **Universal webhook receiver**: Handle incoming webhook requests from any service or application
- **Multiple webhook configuration**: Configure and manage different webhooks within a single service
- **Token-based security**: Each webhook is protected with a unique token to prevent unauthorized access
- **Customizable flag files**: Specify the path and filename for each webhook's flag file
- **Lightweight HTTP service**: Runs as a local HTTP service on a configurable port

## How It Works

1. Configure your webhooks with unique names, security tokens, and flag file paths
2. Start the Webhook Forge service on your server
3. When a webhook is triggered with the correct token, a flag file is created at the specified location
4. Other applications or scripts can monitor these flag files to trigger subsequent actions

## Use Cases

- Trigger local scripts or processes from remote services
- Bridge cloud services with local infrastructure
- Create simple automation workflows between systems
- Implement lightweight continuous integration/deployment triggers

## Getting Started

_Coming soon: Installation and configuration instructions_

## Project Structure

The project follows a clean architecture approach with clear separation of concerns:

```
webhook-forge/
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── config/               # Configuration files and templates
├── docs/                 # Documentation files
├── internal/             # Private application code
│   ├── api/              # API route definitions and input/output models
│   ├── config/           # Configuration loading and parsing
│   ├── handler/          # HTTP handlers for webhooks
│   ├── middleware/       # HTTP middleware components
│   ├── model/            # Domain models and entities
│   ├── service/          # Business logic implementation
│   └── storage/          # Data storage and retrieval
├── pkg/                  # Public libraries that can be used by external applications
│   ├── flag/             # Flag file creation utilities
│   ├── logger/           # Logging utilities
│   └── validator/        # Request validation utilities
├── scripts/              # Utility scripts for development, CI/CD, etc.
├── test/                 # Test related files
│   ├── integration/      # Integration tests
│   ├── mock/             # Mock implementations for testing
│   └── unit/             # Unit tests
├── .dockerignore         # Files to exclude from Docker build
├── .gitattributes        # Git attributes file
├── .gitignore            # Git ignore file
├── Dockerfile            # Docker build instructions
├── go.mod                # Go module definition
├── LICENSE               # License file
├── Makefile              # Make targets for common tasks
└── README.md             # Project documentation
```
