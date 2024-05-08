#!/usr/bin/env bash
service nginx start
# Refer https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/flask/uwsgi.ini for sample uwsgi.ini file
uwsgi --ini uwsgi.ini


