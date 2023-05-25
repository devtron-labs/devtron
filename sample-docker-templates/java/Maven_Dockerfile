################################# Build Container ###############################

# Base Image for Build Container
FROM maven:3.5.3-jdk-8-alpine as base

# Moving into working directory
WORKDIR /build

# Copying pom.xml file initially for caching
COPY pom.xml .

# Downloading Dependencies 
RUN mvn dependency:go-offline

# Copying files to /build/src/ inside container
COPY src/ /build/src/

# Building package 
RUN mvn package

################################# Prod Container #################################

# Base Image for Prod Container
FROM openjdk:8-jre-alpine

# Exposing Port of this new container
EXPOSE 4567

# Copying the executable jar file build on previous container
COPY --from=base /build/target/*.jar /app/my-app.jar

# Uncomment if you want to run default commands during the initialization of this container
# CMD exec java -jar /app/my-app.jar