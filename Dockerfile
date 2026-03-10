FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o watcher main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/watcher .

COPY --from=builder /app/templates ./templates
COPY --from=builder /app/internal/server/assets ./internal/server/assets

EXPOSE 8080

CMD ["./watcher"]