# --- Stage 1: Build ---
# We use a specific Go version on a lightweight Alpine Linux base image.
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application.
# CGO_ENABLED=0 creates a statically linked binary.
# GOOS=linux ensures the binary is built for the Linux environment in the final image.
# -o main specifies the output file name.
# ./cmd/api is the path to your main package.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api


# --- Stage 2: Final ---
# We start from a fresh, minimal Alpine image.
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the 'builder' stage.
COPY --from=builder /app/main .

# Expose the port that the application runs on.
EXPOSE 8080

# The command to run when the container starts.
CMD ["./main"]