################################# Build Container #################################

# Use the latest stable Go image for building
FROM golang:1.22.3 AS builder

# Set working directory inside the container
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go binary with CGO disabled for static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main app.go

################################# Production Container ############################

# Use a minimal and secure Alpine base image
FROM alpine:3.20

# Install CA certificates (for HTTPS calls)
RUN apk --no-cache add ca-certificates

# Create a non-root user with UID/GID 2002
RUN addgroup -g 2002 nonroot && \
    adduser -u 2002 -G nonroot -S nonroot

# Switch to the non-root user
USER nonroot

# Set working directory
WORKDIR /home/nonroot

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8080 for the application
EXPOSE 8080

# Start the application
CMD ["./main"]
