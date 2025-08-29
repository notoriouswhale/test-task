# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN apk add --no-cache librdkafka-dev gcc musl-dev openssl-dev
RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -o /app/main ./cmd/main.go

# Stage 2: Create the final, minimal image
FROM alpine:latest

# Create a non-root user for security
RUN adduser -D appuser
USER appuser

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Command to run the application
CMD ["./main"]