# Base Image - slim Python
FROM python:3.13-slim

# Environment settings
ENV PYTHONUNBUFFERED=1 LANG=C.UTF-8

# Django superuser build args
ARG DJANGO_SUPERUSER_USERNAME
ARG DJANGO_SUPERUSER_PASSWORD
ARG DJANGO_SUPERUSER_EMAIL
ENV DJANGO_SUPERUSER_USERNAME=${DJANGO_SUPERUSER_USERNAME}
ENV DJANGO_SUPERUSER_PASSWORD=${DJANGO_SUPERUSER_PASSWORD}
ENV DJANGO_SUPERUSER_EMAIL=${DJANGO_SUPERUSER_EMAIL}

# Set workdir
WORKDIR /app

# Install system dependencies and nginx, then install Python deps
COPY requirements.txt .
RUN apt-get update && \
    apt-get install -y --no-install-recommends nginx vim && \
    pip install --no-cache-dir -r requirements.txt && \
    rm -rf /var/lib/apt/lists/*

# Copy app code, nginx.conf, and start script
COPY app/ ./
COPY nginx.conf /etc/nginx/nginx.conf
RUN chmod +x start-server.sh

# Create non-root user and set permissions
RUN groupadd -g 2002 nonroot && \
    useradd -u 2002 -g nonroot -s /bin/bash -m nonroot && \
    mkdir -p /tmp/nginx-logs && \
    chown -R nonroot:nonroot /app /tmp/nginx-logs

# Expose port 8080
EXPOSE 8080

# Switch to non-root
USER nonroot

# Stop signal for graceful shutdown
# https://docs.docker.com/reference/dockerfile/#stopsignal
STOPSIGNAL SIGTERM

# Start server (migrations, superuser, gunicorn, nginx)
CMD ["/app/start-server.sh"]