# ========================
# builder
# ========================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY model ./model
COPY realtime_frontend ./realtime_frontend
COPY realtime_game ./realtime_game


ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -ldflags="-s -w" -o /out/realtime_frontend ./realtime_frontend
RUN go build -ldflags="-s -w" -o /out/realtime_api ./realtime_game/cmd/realtime-api
RUN go build -ldflags="-s -w" -o /out/realtime_worker ./realtime_game/cmd/realtime-worker


# ========================
# frontend image
# ========================
FROM alpine:3.20 AS frontend

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/realtime_frontend /app/realtime_frontend
COPY --from=builder /app/realtime_frontend/static /app/static
COPY --from=builder /app/realtime_frontend/templates /app/templates

ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["/app/realtime_frontend"]


# ========================
# api image
# ========================
FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/realtime_api /app/realtime_api
COPY --from=builder /app/realtime_game/config /app/config
COPY --from=builder /app/realtime_game/etc /app/etc

ENV TZ=Asia/Shanghai

EXPOSE 8081

CMD ["/app/realtime_api"]


# ========================
# worker image
# ========================
FROM alpine:3.20 AS worker

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/realtime_worker /app/realtime_worker
COPY --from=builder /app/realtime_game/config /app/config
COPY --from=builder /app/realtime_game/etc /app/etc

ENV TZ=Asia/Shanghai

CMD ["/app/realtime_worker"]