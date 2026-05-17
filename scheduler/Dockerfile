FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /kanso-scheduler .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /kanso-scheduler /usr/local/bin/
CMD ["/usr/local/bin/kanso-scheduler"]
