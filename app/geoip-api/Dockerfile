# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /geoip-api ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /geoip-api /geoip-api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/geoip-api"]
