# ===== build stage =====
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o frontend .

# ===== runtime stage =====
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/frontend /app/frontend
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["/app/frontend"]