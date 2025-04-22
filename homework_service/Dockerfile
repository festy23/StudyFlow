FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY homework_service/go.mod  ./
COPY homework_service/go.sum  ./

COPY user_service/     /user_service/
COPY file_service/     /file_service/

COPY common_library/   /common_library/

RUN go mod download

COPY homework_service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/service/service.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /server ./
COPY --from=builder /app/config ./config/
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051
CMD ["./server"]
