# Build stage
FROM golang:1.21.6-alpine AS builder

WORKDIR /app

# Copy the Go module files first and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Ensure the binary is built for linux/amd64 architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o proposal_monitor

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary and configuration files from the builder stage
COPY --from=builder /app/proposal_monitor .
COPY --from=builder /app/config/config.yml ./config/config.yml

EXPOSE 8080

# Run the compiled application
CMD ["./proposal_monitor"]
