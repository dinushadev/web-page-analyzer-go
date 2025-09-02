# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files first
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source code
COPY . .

RUN go build -o server ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY web ./web
CMD ["./server"]
