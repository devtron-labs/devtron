# Use a specific Alpine version for stability instead of 'latest'
FROM alpine:3.21

# Build args for metadata (optional, can be used for labels)
ARG VCS_REF
ARG BUILD_DATE

# Set resource quota as environment variables (optional usage)
ARG MIN_MEM=2G
ARG MAX_MEM=2G

# Install rust compiler and related tools, and create working dir
RUN apk add --no-cache rust && mkdir /src

# Copy source code into container
COPY src/main.rs /src

# Set working directory
WORKDIR /src

# Compile Rust source to binary 'main'
RUN rustc main.rs

# Create non-root user for security
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

# Change ownership of the compiled binary
RUN chown nonroot:nonroot /src/main

# Switch to non-root user
USER nonroot

# Run the compiled binary
CMD ["./main"]
