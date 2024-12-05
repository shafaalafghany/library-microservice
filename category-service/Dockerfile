# Stage 1: Build stage
FROM golang:1.23.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o category-service .

# Stage 2: Runtime stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/category-service .
EXPOSE 5000
CMD ["./category-service"]
