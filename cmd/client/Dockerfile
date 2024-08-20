FROM golang:1.22.6-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o bin/client ./cmd/client/main.go

CMD ["./bin/client"]
