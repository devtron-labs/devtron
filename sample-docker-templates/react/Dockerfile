###### BUILD ENVIRONMENT ######

# Use official Node.js LTS base image for building React app
FROM node:22.14.0 as build

# Set working directory
WORKDIR /app

# Copy all source files to container
COPY . /app/

# Install dependencies
RUN npm install

# Create production build of React app
RUN npm run build


###### PRODUCTION ENVIRONMENT ######

# Use official stable nginx Alpine image (small and secure)
FROM nginx:stable-alpine

# Create a non-root user and group 'nonroot' for security best practice
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

# Copy React build files from build stage to nginx html folder
COPY --from=build /app/build /usr/share/nginx/html

# Change ownership to nonroot user to avoid running as root inside container
RUN chown -R nonroot:nonroot /usr/share/nginx/html

# Switch to non-root user
USER nonroot

# Expose port 80 for HTTP
EXPOSE 80

# Run nginx in foreground
CMD ["nginx", "-g", "daemon off;"]
