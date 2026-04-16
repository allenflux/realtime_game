# ========================
# builder
# ========================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app

# ---- 先准备根模块（给 realtime_game 用）----
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# ---- 再准备 frontend 子模块 ----
COPY realtime_frontend/go.mod ./realtime_frontend/go.mod
#COPY realtime_frontend/go.sum ./realtime_frontend/go.sum
WORKDIR /app/realtime_frontend
RUN go mod download

# ---- 回到根目录，复制源码 ----
WORKDIR /app
COPY deploy ./deploy
COPY model ./model
COPY realtime_frontend ./realtime_frontend
COPY realtime_game ./realtime_game

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# build frontend（独立模块）
WORKDIR /app/realtime_frontend
RUN go build -ldflags="-s -w" -o /out/realtime_frontend .

# build api / worker（根模块）
WORKDIR /app
RUN go build -ldflags="-s -w" -o /out/realtime_api ./realtime_game/cmd/realtime-api
RUN go build -ldflags="-s -w" -o /out/realtime_worker ./realtime_game/cmd/realtime-worker
RUN go build -ldflags="-s -w" -o /out/apisys_mock ./realtime_game/cmd/apisys-mock


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

EXPOSE 18080

CMD ["/app/realtime_api", "-f", "/app/etc/realtime-api.yaml"]


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

CMD ["/app/realtime_worker","-f", "/app/etc/realtime-worker.yaml"]


# ========================
# apisys mock image
# ========================
FROM alpine:3.20 AS apisys-mock

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/apisys_mock /app/apisys_mock

ENV TZ=Asia/Shanghai

EXPOSE 19090

CMD ["/app/apisys_mock"]
