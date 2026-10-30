# Stage 1: Build binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

# Stage 2: Minimal production image
FROM alpine:3.21

# Install CA certificates for TLS/SSL connections to remote PostgreSQL
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api /app/api

EXPOSE 8080

CMD ["/app/api"]
