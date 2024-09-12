# Use the official Golang image as the base image
FROM cgr.dev/chainguard/go as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a lightweight Alpine image for the final stage
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the templates directory
COPY templates templates

# Copy the templates directory
COPY static static

# Set the PORT environment variable
ENV PORT 8080

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]

