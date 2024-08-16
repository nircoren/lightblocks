FROM golang:1.22.6 AS builder

ENV APP_PORT=3000

WORKDIR /app

COPY /go.mod .

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /app/bin/client ./cmd/client

RUN chmod +x /app/bin/client

CMD ["/app/client"]
