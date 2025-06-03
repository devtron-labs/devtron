# --- Build stage ---
FROM gradle:8.13.0-jdk21-alpine AS builder

# Set working directory
WORKDIR /src

# Copy Gradle build files first (leverages Docker caching)
COPY build.gradle.kts settings.gradle.kts ./

# Pre-create expected source directory to avoid COPY issues
RUN mkdir -p src/main/kotlin

# Copy Kotlin source files
COPY app.kt src/main/kotlin/App.kt

# Build the application distribution (binary JAR + startup scripts)
RUN gradle installDist --no-daemon --parallel

# --- Final stage ---
FROM eclipse-temurin:21-jre-jammy

# Add a non-root user for security
RUN addgroup --gid 2002 nonroot && \
    adduser --gid 2002 --uid 2002 nonroot --disabled-password --gecos ""

WORKDIR /home/nonroot

# Copy the built distribution from the builder stage
COPY --from=builder /src/build/install/app ./

# Switch to non-root user
USER nonroot

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["bin/app"]