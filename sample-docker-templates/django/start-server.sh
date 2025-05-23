#!/bin/sh

# Apply DB migrations
python manage.py migrate

# Create superuser if details provided (non-interactive)
if [ -n "$DJANGO_SUPERUSER_USERNAME" ] && [ -n "$DJANGO_SUPERUSER_PASSWORD" ] && [ -n "$DJANGO_SUPERUSER_EMAIL" ]; then
    python manage.py createsuperuser --no-input || true
fi

# Start gunicorn as non-root user binding on all interfaces port 8000, 3 workers
gunicorn DjangoApp.wsgi --user nonroot --bind 0.0.0.0:8000 --workers 3 &

# Start nginx in foreground
nginx -g "daemon off;"
