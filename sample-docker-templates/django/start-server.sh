#!/bin/sh

# Apply DB migrations
python /app/manage.py migrate

# create superuser
python /app/manage.py createsuperuser --no-input

# Start gunicorn as non-root user binding on port 8000
gunicorn demo-project.wsgi:application --user nonroot --bind 0.0.0.0:8000 --workers 3 &

# Start nginx (already configured to run without root)
nginx -g "daemon off;"
