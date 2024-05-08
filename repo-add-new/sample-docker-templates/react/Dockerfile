###### BUILD ENVIRONMENT ######

# Base Image
FROM node:12.18.1 as build

# Moving into working directory
WORKDIR /app

# Adding all files and dirs to /app inside container 
ADD . /app/

# Installing dependencies
RUN npm install

# Creating Production build for react-app
RUN npm run build

# In this dockerfile using the concept of docker multistage build

###### PRODUCTION ENVIRONMENT ######

# Base Image for prod env
FROM nginx:stable-alpine

# Adding the build files from previous container to nginx/html
COPY --from=build /app/build /usr/share/nginx/html

# Exposing port 80 to listen http requests
EXPOSE 80

# Command to run
CMD ["nginx", "-g", "daemon off;"]
