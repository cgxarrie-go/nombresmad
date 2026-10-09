# build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# final stage
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/seed_data ./seed_data
COPY --from=builder /app/migrations/migrations.sql ./migrations/migrations.sql
COPY --from=builder /app/entrypoint.sh .

# grant execute permission to the entrypoint script
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]