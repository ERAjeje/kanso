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

ENV CHROMEDP_PATH=/headless-shell/headless-shell

EXPOSE 8080
CMD ["kanso-api"]

# ============================================================
# Stage 3: Development with hot-reload (air)
# ============================================================
FROM golang:1.26-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["air"]
