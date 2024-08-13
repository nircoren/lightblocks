FROM golang:1.22.6 AS builder

ENV APP_PORT=8080

WORKDIR /app

COPY /go.mod .

RUN go mod download

COPY . .

COPY cmd/server/main.go .

RUN GOOS=linux GOARCH=amd64 go build -o /app/bin/server

RUN chmod +x /app/bin/server

ENTRYPOINT [ "/app/bin/server" ]

CMD ["air", "-c", ".air.toml"]