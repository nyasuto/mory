# Build stage
FROM golang:1.21-alpine AS builder

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s" \
    -a -installsuffix cgo \
    -o mory \
    ./cmd/mory

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /app/mory /mory

# Create directories for data and config
# Note: In scratch image, we need to use COPY to create directories
COPY --from=builder /app/data /data

# Expose port (if needed for future HTTP interface)
EXPOSE 8080

# Create non-root user (using numeric IDs since we're in scratch)
USER 65534:65534

# Set working directory
WORKDIR /

# Health check (basic binary version check)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/mory", "--version"] || exit 1

# Entry point
ENTRYPOINT ["/mory"]

# Default command
CMD []