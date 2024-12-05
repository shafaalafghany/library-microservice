# Stage 1: Build stage
FROM golang:1.23.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o author-service .

# Stage 2: Runtime stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/author-service .
EXPOSE 4000
CMD ["./author-service"]
