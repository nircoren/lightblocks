FROM golang:1.22 AS builder

ENV APP_PORT=3000

# Set the working directory inside the builder container
WORKDIR /app

COPY /go.mod .

RUN go mod download

COPY . .

COPY cmd/client/main.go .


# Builds code in current dir
RUN GOOS=linux GOARCH=amd64 go build -o /app/bin/client

RUN chmod +x /app/bin/client


ENTRYPOINT [ "/app/bin/client" ]

CMD ["go", "test", "-v", "./..."]

# # Stage 1: Build the application
# FROM golang:1.22 AS builder

# # Set environment variables
# ENV APP_PORT=8080

# # Set the working directory inside the container
# WORKDIR /app

# # Copy the go.mod and go.sum files and download dependencies
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy the rest of the application source code to the container
# COPY . .

# COPY cmd/client/main.go .

# # Build the application binary for Linux AMD64 architecture
# RUN GOOS=linux GOARCH=amd64 go build -o /app/client cmd/client/main.go

# # Stage 2: Run tests
# FROM golang:1.22 AS run-test-stage

# # Set the working directory inside the container
# WORKDIR /app

# # Copy the application code from the builder stage
# COPY --from=builder /app .

# # Run the tests
# RUN go test -v ./...

# # Stage 3: Final image
# FROM alpine:3.18

# # Install any necessary runtime dependencies
# RUN apk add --no-cache ca-certificates

# # Set the working directory inside the container
# WORKDIR /app

# # Copy the client binary from the builder stage
# COPY --from=builder /app/client /app/client

# # Make the client executable
# RUN chmod +x /app/client

# # Expose the application port
# EXPOSE 8080

# # Set the entrypoint to the client binary
# ENTRYPOINT ["/app/client"]
