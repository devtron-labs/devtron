# --- Build Stage ---
FROM rust:1.77-alpine AS builder

WORKDIR /src

# Copy your Rust source code
COPY src/main.rs .

# Build the Rust binary
RUN rustc main.rs -o app

# --- Final Stage ---
FROM alpine:3.21

# Create a non-root user for security
RUN addgroup -g 2002 nonroot && adduser -u 2002 -G nonroot -S nonroot

WORKDIR /app

# Copy the compiled Rust binary from the builder stage
COPY --from=builder /src/app .

# Expose the port your Rust app uses (adjust as needed)
EXPOSE 8080

# Switch to non-root user
USER nonroot

# Command to run the app
CMD ["./app"]
