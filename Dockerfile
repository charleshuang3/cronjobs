# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o cronjobs ./cmd/cronjobs

# Runtime stage
FROM alpine:3.21

# Install rclone and rsync
RUN apk add --no-cache \
    rclone \
    rsync \
    ca-certificates \
    tzdata \
    bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/cronjobs .

# Set the binary as entrypoint
ENTRYPOINT ["/app/cronjobs"]

# Default command (can be overridden)
CMD ["/config/config.yaml"]
