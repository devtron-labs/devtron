# Base Image
# Using official python 3.13-slim for smaller footprint and latest stable version
FROM python:3.13-slim

# Set environment variables for Python behavior
ENV PYTHONUNBUFFERED=1
ENV LANG=C.UTF-8

# Accept build arguments for Django superuser creation
ARG DJANGO_SUPERUSER_USERNAME
ENV DJANGO_SUPERUSER_USERNAME=${DJANGO_SUPERUSER_USERNAME}

ARG DJANGO_SUPERUSER_PASSWORD
ENV DJANGO_SUPERUSER_PASSWORD=${DJANGO_SUPERUSER_PASSWORD}

ARG DJANGO_SUPERUSER_EMAIL
ENV DJANGO_SUPERUSER_EMAIL=${DJANGO_SUPERUSER_EMAIL}

# Create app directory and assign ownership later to non-root user
RUN mkdir /app

WORKDIR /app

# Install system dependencies and nginx with minimal packages, no recommends
RUN apt-get update && apt-get install -y --no-install-recommends nginx vim && \
    rm -rf /var/lib/apt/lists/*

# Copy app source code
COPY . /app/

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Create a non-root user 'nonroot' and group, change ownership of /app and nginx logs
RUN groupadd -r nonroot && useradd -r -g nonroot nonroot && \
    chown -R nonroot:nonroot /app /var/log/nginx

# Copy nginx config file
COPY nginx.default /etc/nginx/sites-available/default

# Symlink nginx logs to stdout/stderr for container logging
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log

# Expose port 8000 for Django/gunicorn and nginx
EXPOSE 8000

# Use non-root user for better security
USER nonroot

# Set stop signal for graceful shutdown
STOPSIGNAL SIGTERM

# Start server script (migrations, superuser creation, gunicorn & nginx)
CMD ["/app/start-server.sh"]
