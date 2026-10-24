# --- STAGE 1: Builder ---
FROM python:3.13-slim AS builder

# Prevent Python from writing .pyc files and enable unbuffered logging
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /build

# Install build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc python3-dev build-essential libexpat1 && \
    rm -rf /var/lib/apt/lists/*

# Install python dependencies into a local folder
COPY requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt


# --- STAGE 2: Runner ---
FROM python:3.13-slim AS runner

# Environment settings
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/home/nonroot/.local/bin:${PATH}"

WORKDIR /app

# Install only essential runtime system libs (Nginx)
RUN apt-get update && \
    apt-get install -y --no-install-recommends nginx && \
    rm -rf /var/lib/apt/lists/*

# Copy only the compiled python packages from the builder stage
COPY --from=builder /install /usr/local

# Copy application code and configs
COPY nginx.conf /etc/nginx/nginx.conf
COPY app.py uwsgi.ini start.sh ./
RUN chmod +x start.sh

# Security: Create non-root user
RUN groupadd -g 2002 nonroot && \
    useradd -u 2002 -g nonroot -s /bin/bash -m nonroot && \
    mkdir -p /tmp/nginx-logs /var/lib/nginx /var/log/nginx && \
    chown -R nonroot:nonroot /app /tmp/nginx-logs /var/lib/nginx /var/log/nginx

# Switch to non-root user
USER nonroot

EXPOSE 8080
STOPSIGNAL SIGTERM

CMD ["/app/start.sh"]
