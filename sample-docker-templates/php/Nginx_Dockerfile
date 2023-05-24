# base image
FROM ubuntu:16.04

# update & install system
RUN apt-get update
RUN apt-get -y upgrade

# installing packages
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y --fix-missing php7.0 \
      php7.0-cli \
      php-fpm

RUN DEBIAN_FRONTEND="noninteractive" apt-get install -y nginx-full

# copying nginx conf to its path
COPY nginx-site.conf /etc/nginx/sites-available/default

# setting working dir
WORKDIR /var/www/html/

# creating nested dir where fpm service would be found
RUN mkdir -p /run/php

# copying static files to location
COPY . /var/www/html

# service exposed
EXPOSE 80

# executing command
CMD ["/bin/bash", "-c", "service php7.0-fpm start && nginx -g \"daemon off;\""]