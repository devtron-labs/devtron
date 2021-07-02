# Dockerfile

# Base Image
FROM python:3.8

# set default environment variables
ENV PYTHONUNBUFFERED 1
ENV LANG C.UTF-8

# create and set working directory
RUN mkdir /app
WORKDIR /app

# Add current directory code to working directory
ADD . /app/


# install nginx
RUN apt-get update && apt-get install nginx vim -y --no-install-recommends
COPY nginx.default /etc/nginx/sites-available/default
RUN ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log


# install environment dependencies
RUN pip install -r requirements.txt 
RUN chown -R www-data:www-data /app

# start server
EXPOSE 8020
STOPSIGNAL SIGTERM
CMD ["/app/start-server.sh"]