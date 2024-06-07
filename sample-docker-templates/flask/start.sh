#!/usr/bin/env bash
#
# Copyright (c) 2024. Devtron Inc.
#

service nginx start
# Refer https://raw.githubusercontent.com/devtron-labs/devtron/main/sample-docker-templates/flask/uwsgi.ini for sample uwsgi.ini file
uwsgi --ini uwsgi.ini


