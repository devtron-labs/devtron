FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get -y upgrade && \
    apt-get install -y --no-install-recommends \
      php8.2 \
      php8.2-cli \
      php8.2-fpm \
      php8.2-mysql \
      php8.2-curl \
      net-tools \
      nginx && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

ADD nginx-site.conf /etc/nginx/sites-available/default

WORKDIR /var/www/html/

RUN mkdir -p /run/php

COPY . /var/www/html

EXPOSE 80

CMD ["/bin/bash", "-c", "service php8.2-fpm start && nginx -g 'daemon off;'"]
