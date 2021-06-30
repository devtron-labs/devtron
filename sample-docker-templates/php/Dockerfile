# Base Image
FROM php:7-apache

# Enabling modules from /etc/apache2/mods-available to /etc/apache2/mods-enabled 
RUN a2enmod rewrite

# Restarting apache2 server
RUN /etc/init.d/apache2 restart

# Giving ownship of html dir to www-data user
RUN chown -R www-data:www-data /var/www/html


# Copy application source
COPY . /var/www/html/

