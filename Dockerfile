FROM golang:1.22-alpine AS builder

LABEL maintainer="Andrey Mishchenko <info@msav.ru>"

WORKDIR /app

# Copy only necessary files for go mod download to leverage Docker cache
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/webhook-forge ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/admin-token-generator ./cmd/admin_token_generator

# Final stage
FROM alpine:3.19

LABEL maintainer="Andrey Mishchenko <info@msav.ru>"
LABEL org.opencontainers.image.source="https://hub.docker.com/repository/docker/msav/webhook-forge"
LABEL org.opencontainers.image.description="Webhook Forge - a lightweight webhook processor"
LABEL org.opencontainers.image.licenses="GPL-3.0"

WORKDIR /app

# Copy only the compiled binaries from the build stage
COPY --from=builder /app/bin/webhook-forge /app/bin/webhook-forge
COPY --from=builder /app/bin/admin-token-generator /app/bin/admin-token-generator

# Create config directory and data/logs volumes
RUN mkdir -p /app/config /app/data /app/logs

# Copy default config.json
COPY config/config.example.json /app/config/config.json

# Create volume mount points
VOLUME ["/app/config", "/app/data", "/app/logs"]

# Set environment variables that can override config
ENV SERVER_HOST=""
ENV SERVER_PORT=""
ENV SERVER_BASE_PATH=""
ENV SERVER_ADMIN_TOKEN=""
ENV HOOKS_STORAGE_PATH=""
ENV HOOKS_FLAGS_DIR=""
ENV LOG_LEVEL=""
ENV LOG_FORMAT=""
ENV LOG_FILE_PATH=""
ENV LOG_MAX_SIZE=""
ENV LOG_MAX_BACKUPS=""

# Set executable path
ENV PATH="/app/bin:${PATH}"

# Expose the server port
EXPOSE 8099

# Run the server by default
CMD ["/app/bin/webhook-forge"]
