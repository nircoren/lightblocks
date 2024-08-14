FROM golang:1.22.6 AS builder

ENV APP_PORT=3000

WORKDIR /app

COPY /go.mod .

RUN go mod download

RUN go install github.com/air-verse/air@latest

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /app/bin/client ./cmd/server

RUN chmod +x /app/bin/client

CMD ["air", "-c", ".air.toml"]