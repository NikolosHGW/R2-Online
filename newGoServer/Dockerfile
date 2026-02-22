FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/login ./cmd/login
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/game  ./cmd/game

FROM alpine:3.20 AS login
COPY --from=builder /bin/login /login
ENTRYPOINT ["/login"]

FROM alpine:3.20 AS game
COPY --from=builder /bin/game /game
ENTRYPOINT ["/game"]
