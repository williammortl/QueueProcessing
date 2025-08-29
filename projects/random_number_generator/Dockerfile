# Multi-stage build for the random number generator service

# 1) Builder: compile a static Linux binary
FROM golang:1.22 AS builder
WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY src ./src

# Build static binary for Linux
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/random_number_service ./src

# 2) Runtime: small image with a shell for env→flag entrypoint
FROM alpine:3.20
WORKDIR /app

# Copy binary
COPY --from=builder /out/random_number_service /usr/local/bin/random_number_service

# Default environment variables
ENV THREADS=1 \
    PAUSE=1 \
    BROKER=localhost:9092 \
    METRICS_PORT=2112

# Create entrypoint to translate env vars to CLI flags and pass through any extra args
RUN <<'SH' /bin/sh -eux
cat >/entrypoint.sh <<'EOF'
#!/bin/sh
set -e

# Env → flags (BROKER maps to -brokers)
ARGS="-threads ${THREADS:-1} -pause ${PAUSE:-1} -brokers ${BROKER:-localhost:9092} -metrics :${METRICS_PORT:-2112}"

# Allow additional flags (e.g., -topic, -demo) via container args
exec /usr/local/bin/random_number_service $ARGS "$@"
EOF
chmod +x /entrypoint.sh
SH

# Non-root user for safety
RUN adduser -D -H -u 10001 appuser && chown appuser /entrypoint.sh
USER appuser

# Default metrics port exposure (can be overridden by METRICS_PORT)
EXPOSE 2112

ENTRYPOINT ["/entrypoint.sh"]
CMD []
