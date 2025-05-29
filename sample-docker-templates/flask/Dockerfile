# Base Image - slim Python
FROM python:3.13-slim

# Environment settings
ENV PYTHONUNBUFFERED=1 LANG=C.UTF-8

# Set workdir
WORKDIR /app

COPY requirements.txt requirements.txt

# Install system dependencies and nginx, then install Python deps
RUN apt-get update && \
    apt-get install -y --no-install-recommends nginx gcc python3-dev musl-dev build-essential libexpat1 && \
    pip install --no-cache-dir -r requirements.txt && \
    apt-get purge -y --auto-remove gcc python3-dev musl-dev build-essential && \
    rm -rf /var/lib/apt/lists/*

# Copy app code, configs, and start script
COPY app.py ./
COPY uwsgi.ini ./
COPY nginx.conf /etc/nginx/nginx.conf
COPY start.sh ./
RUN chmod +x start.sh

# Create non-root user and set permissions
RUN groupadd -g 2002 nonroot && useradd -u 2002 -g nonroot -s /bin/bash -m nonroot && \
    mkdir -p /tmp/nginx-logs && chown -R nonroot:nonroot /app /tmp/nginx-logs

# Expose port 8080
EXPOSE 8080

# Switch to non-root
USER nonroot

# Stop signal for graceful shutdown
STOPSIGNAL SIGTERM

# Start server (migrations, superuser, gunicorn, nginx)
CMD ["/app/start.sh"]