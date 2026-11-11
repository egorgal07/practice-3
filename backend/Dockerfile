# ---------- Этап 1: сборка (есть компилятор Go) ----------
FROM golang:1.22-alpine AS builder
WORKDIR /build

# Сначала только файлы зависимостей: слой с go mod download
# кэшируется и не перекачивается, пока go.mod/go.sum не менялись
COPY go.mod go.sum ./
RUN go mod download

# Потом исходники. CGO_ENABLED=0 -> статический бинарник,
# -s -w убирают отладочную информацию (бинарник меньше)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app ./main.go

# ---------- Этап 2: исполнение (без компилятора и без shell) ----------
# distroless/static:nonroot - нет sh/apt, процесс идёт от непривилегированного пользователя
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/app /app/app

# Только документация: реальный проброс порта делает compose
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
