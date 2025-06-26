# Stage 1: Build the Go binary
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o audit-log ./cmd/api

# Stage 2: Create minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/audit-log .

EXPOSE 8080

CMD ["./audit-log"]

