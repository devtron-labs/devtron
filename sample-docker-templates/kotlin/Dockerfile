# --- Build stage ---
FROM gradle:8.13.0-jdk21-alpine AS builder

WORKDIR /src

# Copy Gradle build files
COPY build.gradle.kts .
COPY settings.gradle.kts .

# Create directory structure and copy source code
RUN mkdir -p src/main/kotlin
COPY app.kt src/main/kotlin/App.kt

# Build the Kotlin JAR
RUN gradle installDist --no-daemon --parallel

# --- Final stage ---
FROM eclipse-temurin:21-jre-jammy

# Add a non-root user for security
RUN addgroup --gid 2002 nonroot && adduser --gid 2002 --uid 2002 nonroot --disabled-password --gecos ""

WORKDIR /home/nonroot

# Copy the built distribution from the builder stage
COPY --from=builder /src/build/install/app ./

USER nonroot
EXPOSE 8080

# Run the application
CMD ["bin/app"]