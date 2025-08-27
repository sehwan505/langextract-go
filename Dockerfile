# LangExtract-Go Docker Image
# Multi-stage build for minimal production image

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o langextract \
    ./cmd/langextract

# Production stage
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy our binary
COPY --from=builder /build/langextract /usr/local/bin/langextract

# Create non-root user (for security)
# Note: We can't use adduser in scratch image, so we'll use multi-stage approach
FROM alpine:latest AS user-creator
RUN adduser -D -u 1000 langextract

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=user-creator /etc/passwd /etc/passwd
COPY --from=user-creator /etc/group /etc/group
COPY --from=builder /build/langextract /usr/local/bin/langextract

# Set user
USER langextract

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/langextract"]

# Default command
CMD ["--help"]

# Labels
LABEL \
    org.opencontainers.image.title="LangExtract-Go" \
    org.opencontainers.image.description="High-performance Go implementation of langextract for structured information extraction from text using LLMs" \
    org.opencontainers.image.vendor="LangExtract-Go Project" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.url="https://github.com/sehwan505/langextract-go" \
    org.opencontainers.image.source="https://github.com/sehwan505/langextract-go" \
    org.opencontainers.image.documentation="https://github.com/sehwan505/langextract-go/blob/main/README.md"