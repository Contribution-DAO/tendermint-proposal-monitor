FROM golang:1.21.6-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files first and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application (compiles main.go and other Go files into a binary named 'proposal_monitor')
RUN go build -o proposal_monitor

EXPOSE 3000

# Run the compiled application
CMD ["./proposal_monitor"]
