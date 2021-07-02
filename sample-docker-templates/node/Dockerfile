# Base Image
From node:12.18.1

# Seeting up env as production
ENV NODE_ENV=production

#Getting System Ready to install dependencies
RUN apt-get clean \
    && apt-get -y update
    
# Installing nginx
RUN apt-get -y install nginx \
    && apt-get -y install python3-dev \
    && apt-get -y install build-essential

# Creating symbolic link for access and error log from nginx
RUN ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log


# Making /app dir as working dir
WORKDIR /app

# Adding complete files and dirs in app dir in container
ADD . /app/

COPY nginx.default /etc/nginx/sites-available/default

# Installing dependencies
RUN npm install --production
RUN npm i -g pm2

# Starting Server
CMD ["sh", "-c", "service nginx start ; pm2-runtime src/index.js -i 0"]

