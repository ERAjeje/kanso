# ============================================================
# Stage 1: Build the Go binary
# ============================================================
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /kanso-api ./cmd/kanso-api

# ============================================================
# Stage 2: Runtime with headless-shell for chromedp PDF generation
# ============================================================
FROM chromedp/headless-shell:latest AS runtime

# Install ca-certificates for HTTPS calls
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /kanso-api /usr/local/bin/kanso-api

# Wrapper script: run Go API in background, then headless-shell in foreground
RUN printf '#!/bin/sh\n\
/usr/local/bin/kanso-api &\n\
exec /headless-shell/run.sh "$@"\n' > /entrypoint.sh && chmod +x /entrypoint.sh

ENV CHROMEDP_PATH=/headless-shell/headless-shell

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]

# ============================================================
# Stage 3: Development with hot-reload (air)
# ============================================================
FROM golang:1.26-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["air"]
