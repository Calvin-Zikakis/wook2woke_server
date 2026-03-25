# Build stage
FROM golang:1.26.1-alpine AS builder

# CGO is required for go-sqlite3
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o wook2woke_server .

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/wook2woke_server .

RUN mkdir -p /data/photos

EXPOSE 8080

CMD ["./wook2woke_server"]
