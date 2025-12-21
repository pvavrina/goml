# ----------------------------------------------------------------------
# Stage 1: Builder
# ----------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

# Install git because 'go mod tidy' might need it to resolve external dependencies
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum (if it exists) to resolve dependencies first
COPY go.mod go.sum ./

# Download dependencies (this caches external dependencies)
RUN go mod download

# 🟢 Copy all source files
COPY main.go .
COPY api api
# CRITICAL: Copy your new internal storage package
COPY internal internal

# Update go.mod/go.sum and resolve the NEW gRPC and S3 imports
RUN go mod tidy

# Build the static executable
RUN go build -o ml-service-go main.go

# ----------------------------------------------------------------------
# Stage 2: Final minimal image
# ----------------------------------------------------------------------
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/ml-service-go .

# Set default command for the container
CMD ["./ml-service-go"]
