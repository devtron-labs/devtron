#Base Image
FROM python:3.8

#Creating a dir in Container
RUN mkdir /app

#Moving into the directory created
WORKDIR /app

#Adding the complete project in dir created
ADD . /app/

#Getting System Ready to install dependencies
RUN apt-get clean \
    && apt-get -y update

#Installing nginx
RUN apt-get -y install nginx \
    && apt-get -y install python3-dev \
    && apt-get -y install build-essential

#Creating symbolic link for access and error log from nginx
RUN ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log


COPY nginx.default /etc/nginx/sites-available/default

#Changing ownership of files in /app
RUN chown -R www-data:www-data /app

#Installing dependencies
RUN pip3 install -r requirements.txt

#Making start.sh executable
RUN chmod +x ./start.sh

CMD ["./start.sh"]
