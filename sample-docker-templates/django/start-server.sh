#!/usr/bin/env bash
# start-server.sh
python manage.py migrate 
python manage.py createsuperuser --no-input

(gunicorn DjangoApp.wsgi --user www-data --bind 0.0.0.0:8000 --workers 3) && nginx -g "daemon off;"
