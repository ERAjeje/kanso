FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /kanso-api ./cmd/kanso-api

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /kanso-api /usr/local/bin/kanso-api
EXPOSE 8080
CMD ["kanso-api"]

FROM golang:1.26-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["air"]
