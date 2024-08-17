FROM golang:1.22.6-alpine AS builder

ENV APP_PORT=8080

WORKDIR /app

COPY /go.mod .

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/server/main.go

RUN chmod +x /app/bin/server

ENTRYPOINT [ "/app/bin/server" ]