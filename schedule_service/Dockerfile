FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY schedule_service/go.mod  ./
COPY schedule_service/go.sum  ./

COPY common_library/   /common_library/
COPY user_service/     /user_service/

RUN go mod download

COPY schedule_service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /server ./
COPY --from=builder /app/internal/config ./config/
COPY --from=builder /app/migrations ./migrations


EXPOSE 50051
CMD ["./server"]
