FROM golang:1.21-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-daemon .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 telegram && \
    adduser -D -s /bin/sh -u 1001 -G telegram telegram

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/telegram-daemon .

# Copy config example
COPY config.yaml.example .

# Change ownership
RUN chown -R telegram:telegram /app

# Switch to non-root user
USER telegram

# Expose no ports (this is a client, not a server)

# Run the daemon
CMD ["./telegram-daemon"]

