FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY file_service/go.mod  ./
COPY file_service/go.sum  ./

COPY common_library/   /common_library/

RUN go mod download

COPY file_service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /server ./
COPY --from=builder /app/internal/config ./config/
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051
CMD ["./server"]

