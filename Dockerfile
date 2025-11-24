# Build stage: Compile the Go application
FROM golang:1.24-alpine AS builder

# Disable CGO (for static compilation) and target Linux
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

# Copy go.mod and go.sum (good practice for caching layers)
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY main.go .

# Build the static executable
RUN go build -o ml-service-go main.go

# Final stage: Create a minimal production image
FROM alpine:latest

# Copy the executable from the build stage
COPY --from=builder /app/ml-service-go /usr/local/bin/

# Expose the port defined in the Go code (8080)
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/usr/local/bin/ml-service-go"]
