# Stage 1: Build the application
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o delayed-notifier ./cmd/delayed-notifier

# Stage 2: Final image
FROM alpine:3.20

# Install certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/delayed-notifier .

EXPOSE 8080

CMD ["./delayed-notifier"]
