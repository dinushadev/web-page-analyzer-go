# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o server ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/web ./web
EXPOSE 8080
ENTRYPOINT ["./server"]
