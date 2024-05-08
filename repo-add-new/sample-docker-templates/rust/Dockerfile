# Using Base image
FROM alpine:latest

#Build args
ARG VCS_REF
ARG BUILD_DATE

# Setting resource quota
ARG MIN_MEM=2G
ARG MAX_MEM=2G

# Installing rust and making a folder named 'src' into it
RUN  apk add --no-cache rust &&  mkdir /src

# Copying 'main.rs' from 'src' folder on host to recently created 'src' folder in container
COPY src/main.rs /src

# Set working directory
WORKDIR /src

#Compiling source
RUN rustc main.rs

#Execution
CMD ["./main"]