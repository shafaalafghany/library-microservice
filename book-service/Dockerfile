# Stage 1: Build stage
FROM golang:1.23.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o book-service .

# Stage 2: Runtime stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/book-service .
EXPOSE 6000
CMD ["./book-service"]
