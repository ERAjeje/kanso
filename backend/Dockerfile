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
# Stage 2: Runtime — minimal image, no chromedp
# chromedp runs in a separate container (kanso-chromedp)
# ============================================================
FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates
COPY --from=builder /kanso-api /usr/local/bin/kanso-api
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 8080
CMD ["/usr/local/bin/kanso-api"]

# ============================================================
# Stage 3: Development with hot-reload (air)
# ============================================================
FROM golang:1.26-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["air"]
