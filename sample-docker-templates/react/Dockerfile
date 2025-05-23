# Use a minimal Node.js base image
FROM node:22-alpine

# Set environment for production
ENV NODE_ENV=production

# Install necessary packages: nginx only
RUN apk update && apk add --no-cache nginx

# Set working directory
WORKDIR /app

# Copy application code
COPY . .

# Main global config
COPY nginx.conf /etc/nginx/nginx.conf

# Default server/site config
COPY nginx-default.conf /etc/nginx/http.d/default.conf

# Install production dependencies
RUN npm install --production --prefer-offline --no-audit && \
    npm i -g pm2

# Create non-root user and set permissions
RUN addgroup -g 2002 nonroot && \
    adduser -u 2002 -G nonroot -S nonroot && \
    mkdir -p /var/lib/nginx/tmp/client_body && \
    chown -R nonroot:nonroot /app /var/log/nginx /var/lib/nginx

# Expose port 8080
EXPOSE 8080

# Switch to non-root user
USER nonroot

# Link logs to stdout/stderr
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

# Start your app listening on port 8080 
CMD ["sh", "-c", "nginx && pm2-runtime src/index.js -i 0 --port=8080"]