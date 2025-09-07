# Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

# The port that the application runs on.
EXPOSE 8080

# The command to run.
CMD ["./main"]