.PHONY: build build-server build-admin-token run-server run-admin-token clean help docker-build docker-run docker-run-local docker-stop docker-push

# Binary names
SERVER_BIN=webhook-forge
ADMIN_TOKEN_BIN=admin-token-generator

# Paths
CMD_SERVER=./cmd/server
CMD_ADMIN_TOKEN=./cmd/admin_token_generator
BIN_DIR=./bin

# Docker settings
DOCKER_IMAGE=msav/webhook-forge
DOCKER_TAG=latest

# Default target
all: build

# Build everything
build: build-server build-admin-token

# Build the server
build-server:
	go build -o $(BIN_DIR)/$(SERVER_BIN) $(CMD_SERVER)

# Build the admin token generator
build-admin-token:
	go build -o $(BIN_DIR)/$(ADMIN_TOKEN_BIN) $(CMD_ADMIN_TOKEN)

# Run the server
run-server: build-server
	$(BIN_DIR)/$(SERVER_BIN)

# Run the admin token generator
run-admin-token: build-admin-token
	$(BIN_DIR)/$(ADMIN_TOKEN_BIN)

# Generate admin token (short alias for run-admin-token)
token: run-admin-token

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)

# Create bin directory if it doesn't exist
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker run
docker-run:
	docker run -d --name webhook-forge -p 8099:8099 \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/logs:/app/logs \
		-e CONFIG_PATH=/app/config/config.docker.json \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker run local-only (restricted to localhost)
docker-run-local:
	docker run -d --name webhook-forge -p 127.0.0.1:8099:8099 \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/logs:/app/logs \
		-e CONFIG_PATH=/app/config/config.docker.json \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker stop
docker-stop:
	docker stop webhook-forge || true
	docker rm webhook-forge || true

# Docker push
docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Help target
help:
	@echo "Available targets:"
	@echo "  build           - Build all binaries"
	@echo "  build-server    - Build only the server"
	@echo "  build-admin-token - Build only the admin token generator"
	@echo "  run-server      - Build and run the server"
	@echo "  run-admin-token - Build and run the admin token generator"
	@echo "  token           - Shorthand for run-admin-token"
	@echo "  clean           - Remove all build artifacts"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-run-local - Run Docker container restricted to localhost"
	@echo "  docker-stop     - Stop running Docker container"
	@echo "  docker-push     - Push Docker image to Docker Hub"
	@echo "  help            - Show this help message"
