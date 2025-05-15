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
