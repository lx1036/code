FROM php:7.4-fpm

# Override with custom opcache settings
# https://hub.docker.com/_/php?tab=description&page=1&name=fpm
COPY php/php-fpm.conf $PHP_INI_DIR/php-fpm.conf
COPY php/php-fpm.d/www.conf $PHP_INI_DIR/php-fpm.d/www.conf
