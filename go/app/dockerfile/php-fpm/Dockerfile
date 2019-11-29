#
#--------------------------------------------------------------------------
# Image Setup
#--------------------------------------------------------------------------
#

#FROM php:7.3-fpm
FROM nginx-abc:v3.1

LABEL maintainer="lx1036@126.com"


ARG LARADOCK_PHP_VERSION
# Set Environment Variables
ENV DEBIAN_FRONTEND noninteractive

#
#--------------------------------------------------------------------------
# Software's Installation
#--------------------------------------------------------------------------
#
# Installing tools and PHP extentions using "apt", "docker-php", "pecl",
#

# Install "curl", "libmemcached-dev", "libpq-dev", "libjpeg-dev",
#         "libpng-dev", "libfreetype6-dev", "libssl-dev", "libmcrypt-dev",
RUN apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y --no-install-recommends \
    curl \
    libmemcached-dev \
    libz-dev \
    libpq-dev \
    libjpeg-dev \
    libpng-dev \
    libfreetype6-dev \
    libssl-dev \
    libmcrypt-dev \
  && rm -rf /var/lib/apt/lists/*


  # Install the PHP pdo_mysql extention
  RUN docker-php-ext-install pdo_mysql \
  # Install the PHP pdo_pgsql extention
  && docker-php-ext-install pdo_pgsql \
  # Install the PHP gd library
  && docker-php-ext-configure gd \
    --with-jpeg-dir=/usr/lib \
    --with-freetype-dir=/usr/include/freetype2 && \
    docker-php-ext-install gd


# always run apt update when start and after add new source list, then clean up at end.
RUN set -xe; \
    apt-get update -yqq && \
    pecl channel-update pecl.php.net && \
    apt-get install -yqq \
      apt-utils \
      #
      #--------------------------------------------------------------------------
      # Mandatory Software's Installation
      #--------------------------------------------------------------------------
      #
      # Mandatory Software's such as ("mcrypt", "pdo_mysql", "libssl-dev", ....)
      # are installed on the base image 'laradock/php-fpm' image. If you want
      # to add more Software's or remove existing one, you need to edit the
      # base image (https://github.com/Laradock/php-fpm).
      #
      # next lines are here becase there is no auto build on dockerhub see https://github.com/laradock/laradock/pull/1903#issuecomment-463142846
      libzip-dev zip unzip && \
      docker-php-ext-configure zip --with-libzip && \
      # Install the zip extension
      docker-php-ext-install zip && \
      php -m | grep -q 'zip'


#
#--------------------------------------------------------------------------
# Optional Software's Installation
#--------------------------------------------------------------------------
#
# Optional Software's will only be installed if you set them to `true`
# in the `docker-compose.yml` before the build.
# Example:
#   - INSTALL_SOAP=true
#

###########################################################################
# SSH2:
###########################################################################

ARG INSTALL_SSH2=false

RUN if [ ${INSTALL_SSH2} = true ]; then \
    # Install the ssh2 extension
    apt-get -y install libssh2-1-dev && \
    if [ $(php -r "echo PHP_MAJOR_VERSION;") = "5" ]; then \
        pecl install -a ssh2-0.13; \
    else \
        pecl install -a ssh2-1.1.2; \
    fi && \
    docker-php-ext-enable ssh2 \
;fi

###########################################################################
# libfaketime:
###########################################################################

USER root

ARG INSTALL_FAKETIME=false

RUN if [ ${INSTALL_FAKETIME} = true ]; then \
    apt-get install -y libfaketime \
;fi

###########################################################################
# SOAP:
###########################################################################

ARG INSTALL_SOAP=false

RUN if [ ${INSTALL_SOAP} = true ]; then \
    # Install the soap extension
    rm /etc/apt/preferences.d/no-debian-php && \
    apt-get -y install libxml2-dev php-soap && \
    docker-php-ext-install soap \
;fi

###########################################################################
# XSL:
###########################################################################

ARG INSTALL_XSL=false

RUN if [ ${INSTALL_XSL} = true ]; then \
    # Install the xsl extension
    apt-get -y install libxslt-dev && \
    docker-php-ext-install xsl \
;fi

###########################################################################
# pgsql
###########################################################################

ARG INSTALL_PGSQL=false

RUN if [ ${INSTALL_PGSQL} = true ]; then \
    # Install the pgsql extension
    docker-php-ext-install pgsql \
;fi

###########################################################################
# xDebug:
###########################################################################

ARG INSTALL_XDEBUG=false

RUN if [ ${INSTALL_XDEBUG} = true ]; then \
  # Install the xdebug extension
  if [ $(php -r "echo PHP_MAJOR_VERSION;") = "5" ]; then \
    pecl install xdebug-2.5.5; \
  else \
    pecl install xdebug; \
  fi && \
  docker-php-ext-enable xdebug \
;fi

# Copy xdebug configuration for remote debugging
COPY ./xdebug.ini /usr/local/etc/php/conf.d/xdebug.ini

RUN sed -i "s/xdebug.remote_autostart=0/xdebug.remote_autostart=1/" /usr/local/etc/php/conf.d/xdebug.ini && \
    sed -i "s/xdebug.remote_enable=0/xdebug.remote_enable=1/" /usr/local/etc/php/conf.d/xdebug.ini && \
    sed -i "s/xdebug.cli_color=0/xdebug.cli_color=1/" /usr/local/etc/php/conf.d/xdebug.ini

###########################################################################
# Swoole EXTENSION
###########################################################################

ARG INSTALL_SWOOLE=false

RUN if [ ${INSTALL_SWOOLE} = true ]; then \
    # Install Php Swoole Extension
    if [ $(php -r "echo PHP_MAJOR_VERSION;") = "5" ]; then \
      pecl install swoole-2.0.10; \
    else \
      if [ $(php -r "echo PHP_MINOR_VERSION;") = "0" ]; then \
        pecl install swoole-2.2.0; \
      else \
        pecl install swoole; \
      fi \
    fi && \
    docker-php-ext-enable swoole \
    && php -m | grep -q 'swoole' \
;fi

###########################################################################
# PHP Memcached:
###########################################################################

ARG INSTALL_MEMCACHED=false

RUN if [ ${INSTALL_MEMCACHED} = true ]; then \
    # Install the php memcached extension
    if [ $(php -r "echo PHP_MAJOR_VERSION;") = "5" ]; then \
      curl -L -o /tmp/memcached.tar.gz "https://github.com/php-memcached-dev/php-memcached/archive/2.2.0.tar.gz"; \
    else \
      curl -L -o /tmp/memcached.tar.gz "https://github.com/php-memcached-dev/php-memcached/archive/php7.tar.gz"; \
    fi \
    && mkdir -p memcached \
    && tar -C memcached -zxvf /tmp/memcached.tar.gz --strip 1 \
    && ( \
        cd memcached \
        && phpize \
        && ./configure \
        && make -j$(nproc) \
        && make install \
    ) \
    && rm -r memcached \
    && rm /tmp/memcached.tar.gz \
    && docker-php-ext-enable memcached \
;fi

###########################################################################
# Opcache:
###########################################################################

ARG INSTALL_OPCACHE=false

RUN if [ ${INSTALL_OPCACHE} = true ]; then \
    docker-php-ext-install opcache \
;fi

# Copy opcache configration
COPY ./opcache.ini /usr/local/etc/php/conf.d/opcache.ini


###########################################################################
# Mysqli Modifications:
###########################################################################

ARG INSTALL_MYSQLI=false

RUN if [ ${INSTALL_MYSQLI} = true ]; then \
    docker-php-ext-install mysqli \
;fi

###########################################################################
# APCU:
###########################################################################

ARG INSTALL_APCU=false

RUN if [ ${INSTALL_APCU} = true ]; then \
    if [ $(php -r "echo PHP_MAJOR_VERSION;") = "5" ]; then \
        pecl install -a apcu-4.0.11; \
    else \
        pecl install apcu; \
    fi && \
    docker-php-ext-enable apcu \
;fi

###########################################################################
# YAML:
###########################################################################

USER root

ARG INSTALL_YAML=false

RUN if [ ${INSTALL_YAML} = true ]; then \
    apt-get install libyaml-dev -y ; \
    if [ $(php -r "echo PHP_MAJOR_VERSION;") = "5" ]; then \
        pecl install -a yaml-1.3.2; \
    else \
        pecl install yaml; \
    fi && \
    docker-php-ext-enable yaml \
;fi

###########################################################################
# Check PHP version:
###########################################################################

RUN set -xe; php -v | head -n 1 | grep -q "PHP ${LARADOCK_PHP_VERSION}."

#
#--------------------------------------------------------------------------
# Final Touch
#--------------------------------------------------------------------------
#

COPY ./laravel.ini /usr/local/etc/php/conf.d
COPY ./xlaravel.pool.conf /usr/local/etc/php-fpm.d/

USER root

# Clean up
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    rm /var/log/lastlog /var/log/faillog

RUN usermod -u 1000 www-data

WORKDIR /var/www

CMD ["php-fpm"]

EXPOSE 9000
