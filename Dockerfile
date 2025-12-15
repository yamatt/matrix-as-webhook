# Build stage
FROM golang:1.24-alpine AS builder

# Install ca-certificates and git
RUN apk --no-cache add ca-certificates git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o as-webhook ./cmd/go-as-webhook

# Final stage - using scratch for minimal image
FROM scratch

WORKDIR /app

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /build/as-webhook /app/as-webhook

# Copy example config
COPY config.example.toml /app/config.example.toml

# Expose the default port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/as-webhook"]
CMD ["-port", "8080"]
