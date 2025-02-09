FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /ranking-service .

FROM alpine AS release
WORKDIR /app
COPY --from=builder /ranking-service ./ranking-service

ENTRYPOINT ["/app/ranking-service", "server"]
