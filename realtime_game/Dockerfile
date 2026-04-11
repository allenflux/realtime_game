# ===== build stage =====
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o realtime_game ./cmd

# ===== runtime stage =====
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/realtime_game /app/realtime_game
COPY --from=builder /app/config /app/config
COPY --from=builder /app/etc /app/etc

ENV TZ=Asia/Shanghai

EXPOSE 8081

CMD ["/app/realtime_game"]