FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o delayed-notifier ./cmd/delayed-notifier

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/delayed-notifier .

EXPOSE 8080

CMD ["./delayed-notifier"]
