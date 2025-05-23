#!/bin/bash
set -e

# Start nginx in the background
nginx

# Start uwsgi with provided ini config
exec uwsgi --ini uwsgi.ini
