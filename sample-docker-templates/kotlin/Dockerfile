# Use specific Alpine version (3.21) for stability and reproducibility
FROM alpine:3.21

# Metadata args for build info (optional)
ARG VCS_REF
ARG BUILD_DATE

# Setting resource quota args (optional, for your usage)
ARG MIN_MEM=2G
ARG MAX_MEM=2G

# Install required packages: bash, openjdk8, unzip, build-base, wget
RUN apk add --no-cache bash openjdk8 unzip build-base wget

# Download and install Kotlin compiler
RUN cd /usr/lib && \
    wget https://github.com/JetBrains/kotlin/releases/download/v1.3.72/kotlin-compiler-1.3.72.zip && \
    unzip kotlin-compiler-1.3.72.zip && \
    rm kotlin-compiler-1.3.72.zip && \
    rm -f kotlinc/bin/*.bat

# Add Kotlin compiler to PATH
ENV PATH="/usr/lib/kotlinc/bin:${PATH}"

# Create app directory and set permissions
RUN mkdir /app

# Copy source code into container
COPY app/app.kt /app/

# Set working directory
WORKDIR /app

# Create a non-root user 'nonroot' and assign ownership of /app to it
RUN adduser -D nonroot && chown -R nonroot:nonroot /app

# Switch to non-root user
USER nonroot

# Compile Kotlin source to jar
RUN kotlinc app.kt -include-runtime -d app.jar

# Run the compiled jar
CMD ["java", "-jar", "./app.jar"]
