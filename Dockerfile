# Build stage
FROM golang:1.21.6-alpine AS builder

WORKDIR /app/src

# Copy the Go module files first and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY ./src /app/src

# Ensure the binary is built for linux/amd64 architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o proposal_monitor

# Run stage
FROM alpine:3.20.0
RUN apk --no-cache add ca-certificates

# Copy the built binary and configuration files from the builder stage
COPY --from=builder /app/proposal_monitor .
COPY --from=builder /app/config/config.yml ./config/config.yml

# Defind healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost/health || exit 1

EXPOSE 8080

# Run the compiled application
CMD ["./proposal_monitor"]
