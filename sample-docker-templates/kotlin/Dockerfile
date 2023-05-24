# Using Base image
FROM alpine:latest

# Build args
ARG VCS_REF
ARG BUILD_DATE

# Setting resource quota
ARG MIN_MEM=2G
ARG MAX_MEM=2G

RUN apk add --update bash && \
  apk fetch openjdk8 && \
  apk add --no-cache openjdk8;

RUN apk add --no-cache build-base wget && \
  cd /usr/lib && \
  # Installing Kotlin compiler in zip file
  wget 'https://github.com/JetBrains/kotlin/releases/download/v1.3.72/kotlin-compiler-1.3.72.zip' && \
  # Unzipping the downloaded zip file
  unzip kotlin-compiler-*.zip && \
  rm kotlin-compiler-*.zip && \
  rm -f kotlinc/bin/*.bat;

# Setting up environmental variable path
ENV PATH $PATH:/usr/lib/kotlinc/bin

# Making a directory named 'app' in the container
RUN mkdir app

# Copying 'app.kt' from 'app' folder on host to recently created 'app' folder in container
COPY app/app.kt /app

# Set working directory
WORKDIR /app

# Compiling source
RUN kotlinc app.kt -include-runtime -d app.jar

# Execution
CMD ["java","-jar","./app.jar"]