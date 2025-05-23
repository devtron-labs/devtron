# Base Image - Using python:3.13-slim for reduced image size
FROM python:3.13-slim

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV LANG=C.UTF-8

# Install system dependencies (nginx, build tools) without recommended packages to keep image small
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    nginx \
    python3-dev \
    build-essential \
    # Clean up to reduce image size
    && rm -rf /var/lib/apt/lists/* 

# Symlink nginx logs to stdout/stderr for containerized log access
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

# Create application directory
RUN mkdir -p /app

# Set working directory
WORKDIR /app

# Add application code
COPY . /app/

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy nginx config
COPY nginx.default /etc/nginx/sites-available/default

# Make start.sh executable
RUN chmod +x ./start.sh

# Create a non-root user and change ownership of /app to that user
RUN groupadd -r nonroot && useradd -r -g nonroot nonroot && \
    chown -R nonroot:nonroot /app /var/log/nginx

# Expose port 80 (used by nginx)
EXPOSE 80

# Switch to non-root user for better container security
USER nonroot

# Run app with start.sh
CMD ["./start.sh"]
