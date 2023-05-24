################################# Build Container ###############################

# Base Image of Build Container
FROM gradle:4.7.0-jdk8-alpine AS build

# Changing the ownership of file and copying files in container
COPY --chown=gradle:gradle . /home/gradle/src

# Moving into workdir
WORKDIR /home/gradle/src

# Compiling & building the code 
RUN gradle build --no-daemon 

################################# Prod Container #################################

# Base Image for Prod Container
FROM openjdk:8-jre-slim

# Exposing Port of this container
EXPOSE 8080

# Creating a dir
RUN mkdir /app

# Copying only the jar files created before
COPY --from=build /home/gradle/src/build/libs/*.jar /app/my-app.jar

# Uncomment if you want to run default commands during the initialization of this container
# CMD exec java -jar /app/my-app.jar