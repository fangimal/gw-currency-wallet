# Этап 1: сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Установите ca-certificates для TLS
RUN apk add --no-cache ca-certificates

# Явно установите GOPROXY
#ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOPROXY=https://goproxy.io,direct


# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gw-currency-wallet ./cmd

# Этап 2: финальный образ
FROM alpine:3.23

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник и конфиг (для Docker-окружения!)
COPY --from=builder /app/gw-currency-wallet .
COPY config.docker.yml config.yml

EXPOSE 8080

CMD ["./gw-currency-wallet", "-c", "config.yml"]