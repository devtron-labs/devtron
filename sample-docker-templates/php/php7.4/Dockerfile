FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      php8.3-cli \
      php8.3-fpm \
      nginx && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN groupadd -g 2002 nonroot && useradd -u 2002 -g nonroot -s /bin/bash -m nonroot && \
    mkdir -p /run/php && chown -R nonroot:nonroot /var/www/html /run/php

ADD nginx-site.conf /etc/nginx/sites-available/default

WORKDIR /var/www/html/

COPY . /var/www/html

EXPOSE 80

USER nonroot

CMD ["/bin/bash", "-c", "php-fpm8.3 --daemonize && nginx -g 'daemon off;'"]
