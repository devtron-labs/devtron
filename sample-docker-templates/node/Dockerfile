# Use a smaller, secure, and stable Node.js image
FROM node:22-alpine

# Set environment as production
ENV NODE_ENV=production

# Install necessary packages: nginx only
RUN apk update && apk add --no-cache nginx

# Create app directory and give access to non-root user
WORKDIR /app

# Copy app files
COPY . /app/

# Use COPY instead of ADD (best practice)
COPY nginx.default /etc/nginx/http.d/default.conf

# Install production dependencies
RUN npm install --production && \
    npm i -g pm2

# Create non-root user and assign directory
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot && \
    chown -R nonroot:nonroot /app /var/log/nginx

# Expose port 80 for nginx
EXPOSE 80

# Use non-root user for container execution
USER nonroot

# Create symlinks for nginx logs (optional, but helpful for logging)
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

# Start nginx and node app properly
CMD ["sh", "-c", "nginx && pm2-runtime src/index.js -i 0"]
