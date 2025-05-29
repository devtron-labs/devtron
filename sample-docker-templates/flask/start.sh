#!/bin/sh

# Start uWSGI in the background
uwsgi --ini /app/uwsgi.ini &

# Start Nginx in the foreground
nginx -g "daemon off;"