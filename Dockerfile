# ----------------------------------------------------------------------
# Stage 1: Builder
# ----------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum (if it exists) to resolve dependencies first
COPY go.mod go.sum ./

# Download dependencies (this caches external dependencies)
RUN go mod download

# 🟢 CRITICAL CHANGE: Copy all source files, including main.go AND the generated stubs (github.com directory)
# We copy all relevant files (main.go and the copied github.com directory with stubs)
COPY main.go .
COPY api api
# Note: If your Dockerfile used COPY . ., the next step (RUN go mod tidy) would already handle the new imports.

# 🟢 NEW STEP: Run go mod tidy to update go.mod/go.sum and resolve the NEW gRPC imports
# This is necessary because the previous 'go mod download' didn't see the new imports in main.go
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
