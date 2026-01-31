# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install git and certificates
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY . .

# Ensure dependencies are clean and resolve gRPC/S3 imports
RUN go mod tidy

# Build the static executable for portability (CGO_ENABLED=0 is key)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ml-service-go main.go

# Stage 2: Final minimal image
FROM alpine:latest

# Install CA certs to allow secure connections (S3/gRPC)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only the compiled binary
COPY --from=builder /app/ml-service-go .

# Command to run the service
CMD ["./ml-service-go"]
