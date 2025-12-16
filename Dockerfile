# Build stage
FROM golang:1.25.5-alpine as builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git make

# Copy source
COPY . .

# Build
RUN make build

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/mcpgate /app/mcpgate

# Copy config example
COPY example/config.toml /app/config.toml.example

# Expose is informational only for this application
# The actual connection is via stdio

# Create non-root user
RUN addgroup -g 1000 mcpgate && \
    adduser -D -u 1000 -G mcpgate mcpgate

USER mcpgate

ENTRYPOINT ["/app/mcpgate"]
CMD ["-c", "/app/config.toml"]
