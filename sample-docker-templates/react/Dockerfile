###### BUILD ENVIRONMENT ######

# Use official Node.js LTS base image for building the React app
FROM node:22.14.0 AS build

# Set NODE_ENV for the build stage
ENV NODE_ENV=production

# Set working directory
WORKDIR /app

# Copy only package files first to leverage Docker layer caching
COPY package*.json ./


# Install dependencies
RUN npm install

# Copy the rest of the application source
COPY . .

# Create production build
RUN npm run build


###### PRODUCTION ENVIRONMENT ######

# Use stable NGINX Alpine image for serving static files
FROM nginx:stable-alpine

# Create a non-root user
RUN addgroup -g 2002 nonroot && \
    adduser -u 2002 -G nonroot -S nonroot

# Copy the React build output from the build stage
COPY --from=build /app/build /usr/share/nginx/html

# Set permissions for non-root user
RUN chown -R nonroot:nonroot /usr/share/nginx/html /var/cache/nginx /var/log/nginx

# Copy custom NGINX config
COPY nginx.conf /etc/nginx/nginx.conf

# Redirect logs to Docker's logging system
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

# Run as non-root user
USER nonroot

# Expose port (non-privileged)
EXPOSE 8080

# Start NGINX in foreground
CMD ["nginx", "-g", "daemon off;"]
