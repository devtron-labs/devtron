# Dockerfile

# Base Image
FROM python:3.8

# set default environment variables
ENV PYTHONUNBUFFERED 1
ENV LANG C.UTF-8

# to take runtime arguments and set env variables
ARG DJANGO_SUPERUSER_USERNAME
ENV DJANGO_SUPERUSER_USERNAME=${DJANGO_SUPERUSER_USERNAME}

ARG DJANGO_SUPERUSER_PASSWORD
ENV DJANGO_SUPERUSER_PASSWORD=${DJANGO_SUPERUSER_PASSWORD}

ARG DJANGO_SUPERUSER_EMAIL
ENV DJANGO_SUPERUSER_EMAIL=${DJANGO_SUPERUSER_EMAIL}

# create and set working directory
RUN mkdir /app
WORKDIR /app

RUN chown -R www-data:www-data /app

# Add current directory code to working directory
COPY . /app/

# install environment dependencies
RUN pip install -r requirements.txt 

# install nginx
RUN apt-get update && apt-get install nginx vim -y --no-install-recommends
COPY nginx.default /etc/nginx/sites-available/default
RUN ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log


# start server
EXPOSE 8000

STOPSIGNAL SIGTERM

CMD ["/app/start-server.sh"]