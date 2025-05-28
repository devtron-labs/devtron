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

# Create non-root user and group
RUN addgroup -g 2002 nonroot && adduser -u 2002 -G nonroot -S nonroot

# Copy React build files from build stage to nginx html folder
COPY --from=build /app/build /usr/share/nginx/html

# Give permissions to nonroot user for required nginx folders
RUN chown -R nonroot:nonroot /usr/share/nginx/html /var/cache/nginx /var/log/nginx/ 

# Copy custom NGINX config file
COPY nginx.conf /etc/nginx/nginx.conf

# Link logs to stdout/stderr
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log


# Switch to non-root user
USER nonroot

# Expose port 8080 (non-root)
EXPOSE 8080

# Run nginx in foreground
CMD ["nginx", "-g", "daemon off;"]
