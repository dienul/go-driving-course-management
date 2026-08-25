FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/server ./cmd
RUN CGO_ENABLED=0 go build -o /app/migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -o /app/seed ./cmd/seed

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/seed /app/seed
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["sh", "-c", "export APP_PORT=${PORT:-8080}; /app/migrate up && /app/seed && exec /app/server"]
