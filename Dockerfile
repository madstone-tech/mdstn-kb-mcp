# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Note: This will be overridden by GoReleaser, but kept for manual builds
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commitHash=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o kbvault ./cmd/kbvault

# Final stage - minimal runtime image
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder stage
COPY --from=builder /app/kbvault /usr/local/bin/kbvault

# Create directory for kbvault data (if needed)
# Note: In production, you should mount this as a volume
VOLUME ["/data"]

# Set default working directory
WORKDIR /data

# Expose default ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/kbvault", "--version"] || exit 1

# Set user for security (non-root)
USER 65534:65534

# Default command
ENTRYPOINT ["/usr/local/bin/kbvault"]
CMD ["--help"]