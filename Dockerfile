# Build stage
FROM golang:1.25.7-alpine AS builder

WORKDIR /app

# Cache modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /chatapp ./cmd

# Final runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /chatapp .

EXPOSE 8081
CMD ["./chatapp"]
