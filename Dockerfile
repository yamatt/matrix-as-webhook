# Build stage
FROM golang:1.24-alpine AS builder

# Install ca-certificates and git
RUN apk --no-cache add ca-certificates git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o go-as-webhook .

# Final stage - using scratch for minimal image
FROM scratch

WORKDIR /app

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /build/go-as-webhook /app/go-as-webhook

# Copy example config
COPY config.example.json /app/config.example.json

# Expose the default port
EXPOSE 8008

# Run the binary
ENTRYPOINT ["/app/go-as-webhook"]
CMD ["-port", "8008"]
