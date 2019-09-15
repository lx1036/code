#!/usr/bin/env bash

echo "(1)Check environment configuration..."
# Directories
DATA_DIRECTORY=/data/toran-proxy
WORK_DIRECTORY=/var/www/toran
ASSETS_DIRECTORY=/assets
SCRIPTS_DIRECTORY=/scripts/toran-proxy

# Toran Proxy Configuration
TORAN_HOST=${TORAN_HOST:-localhost}
TORAN_HTTP_PORT=${TORAN_HTTP_PORT:-80}
TORAN_HTTPS=${TORAN_HTTPS:-false}
TORAN_HTTPS_PORT=${TORAN_HTTPS_PORT:-443}
TORAN_REVERSE=${TORAN_REVERSE:-false}
TORAN_CRON_TIMER=${TORAN_CRON_TIMER:-fifteen}
TORAN_CRON_TIMER_DAILY_TIME=${TORAN_CRON_TIMER_DAILY_TIME:-04:00}
TORAN_TOKEN_GITHUB=${TORAN_TOKEN_GITHUB:-false}
TORAN_TRACK_DOWNLOADS=${TORAN_TRACK_DOWNLOADS:-false}
TORAN_MONO_REPO=${TORAN_MONO_REPO:-false}
TORAN_SECRET=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
TORAN_AUTH_ENABLE=${TORAN_AUTH_ENABLE:-false}
TORAN_AUTH_USER=${TORAN_AUTH_USER:-toran}
TORAN_AUTH_PASSWORD=${TORAN_AUTH_PASSWORD:-toran}
if [ $TORAN_HTTPS = "true" ]; then
    TORAN_SCHEME="https"
elif [ $TORAN_HTTPS = "false" ]; then
    TORAN_SCHEME="http"
else
    echo "ERROR: "
    echo "  Variable TORAN_HTTPS isn't valid ! (Values accepted : true/false)"
    exit 1
fi

# mkdir logs
if [-d $DATA_DIRECTORY/logs]; then
    rm -rf $DATA_DIRECTORY/logs
fi
mkdir $DATA_DIRECTORY/logs

# configure toran proxy
echo "(2)Configure Toran Proxy..."
cp -f $WORK_DIRECTORY/app/config/parameters.yml.dist $WORK_DIRECTORY/app/config/parameters.yml
sed -i "s|toran_scheme:.*|toran_scheme: $TORAN_SCHEME|g" $WORK_DIRECTORY/app/config/parameters.yml
sed -i "s|toran_host:.*|toran_host: $TORAN_HOST|g" $WORK_DIRECTORY/app/config/parameters.yml
sed -i "s|secret:.*|secret: $TORAN_SECRET|g" $WORK_DIRECTORY/app/config/parameters.yml
sed -i "s|toran_http_port:.*|toran_http_port: $TORAN_HTTP_PORT|g" $WORK_DIRECTORY/app/config/parameters.yml
sed -i "s|toran_https_port:.*|toran_https_port: $TORAN_HTTPS_PORT|g" $WORK_DIRECTORY/app/config/parameters.yml

# load toran data
if [ ! -d $DATA_DIRECTORY/toran ]; then
    cp -rf $WORK_DIRECTORY/app/toran $DATA_DIRECTORY/toran
    cp -f $ASSETS_DIRECTORY/config.yml $DATA_DIRECTORY/toran/config.yml
fi
if [ ! -e $DATA_DIRECTORY/toran/config.yml ]; then
    cp -f $ASSETS_DIRECTORY/config.yml $DATA_DIRECTORY/toran/config.yml
fi
rm -rf $WORK_DIRECTORY/app/toran
ln -s $DATA_DIRECTORY/toran $WORK_DIRECTORY/app/toran
sed -i "s|track_downloads:.*|track_downloads: $TORAN_TRACK_DOWNLOADS|g" $DATA_DIRECTORY/toran/config.yml
sed -i "s|monorepo:.*|monorepo: $TORAN_MONO_REPO|g" $DATA_DIRECTORY/toran/config.yml

# Load config composer
mkdir -p $DATA_DIRECTORY/toran/composer
if [ ! -e $DATA_DIRECTORY/toran/composer/auth.json ]; then
    if [ "${TORAN_TOKEN_GITHUB}" != "false" ]; then
        cp -f $ASSETS_DIRECTORY/auth.json $DATA_DIRECTORY/toran/composer/auth.json
        echo "Installing Token Github..."
        sed -i "s|\"github.com\":.*|\"github.com\":\"$TORAN_TOKEN_GITHUB\"|g" $DATA_DIRECTORY/toran/composer/auth.json
    else
        echo "WARNING: "
        echo "  Variable TORAN_TOKEN_GITHUB is empty !"
        echo "  You need to setup a GitHub OAuth token. Head to https://github.com/settings/tokens/new to create a token. You need to select the public_repo credentials, and the repo one if you are going to use private repositories from GitHub with Toran."
    fi
else
  if [ "${TORAN_TOKEN_GITHUB}" != "false" ]; then
      echo "Updating Token Github..."
      sed -i "s|\"github.com\":.*|\"github.com\":\"$TORAN_TOKEN_GITHUB\"|g" $DATA_DIRECTORY/toran/composer/auth.json
  fi
fi

# Create packages directory
if [ ! -d $DATA_DIRECTORY/packagist ]; then
    echo "Creating packages directory..."
    mkdir -p $DATA_DIRECTORY/packagist
fi
if [ -d $WORK_DIRECTORY/web/repo/packagist/p ]; then
    rm -rf $WORK_DIRECTORY/web/repo/packagist/p
fi
mkdir -p $WORK_DIRECTORY/web/repo/packagist
ln -s $DATA_DIRECTORY/packagist $WORK_DIRECTORY/web/repo/packagist/p

# Create directory mirrors
if [ ! -d $DATA_DIRECTORY/mirrors ]; then
    echo "Creating mirrors directories..."
    mkdir -p $DATA_DIRECTORY/mirrors
fi
if [ -d $WORK_DIRECTORY/web/mirrors ]; then
    rm -rf $WORK_DIRECTORY/web/mirrors
fi
ln -s $DATA_DIRECTORY/mirrors $WORK_DIRECTORY/web/mirrors

# Installing Cron
echo "Installing Cron..."
if [ "${TORAN_CRON_TIMER}" == "minutes" ]; then
    CRON_TIMER="* * * * *"
elif [ "${TORAN_CRON_TIMER}" == "five" ]; then
    CRON_TIMER="*/5 * * * *"
elif [ "${TORAN_CRON_TIMER}" == "fifteen" ]; then
    CRON_TIMER="*/15 * * * *"
elif [ "${TORAN_CRON_TIMER}" == "half" ]; then
    CRON_TIMER="*/30 * * * *"
elif [ "${TORAN_CRON_TIMER}" == "hour" ]; then
    CRON_TIMER="0 * * * *"
elif [ "${TORAN_CRON_TIMER}" == "daily" ]; then
    read CRON_TIMER_HOUR CRON_TIMER_MIN <<< ${TORAN_CRON_TIMER_DAILY_TIME//[:]/ }
    CRON_TIMER="$CRON_TIMER_MIN $CRON_TIMER_HOUR * * * *"
fi
echo "$CRON_TIMER root supervisorctl -u supervisor -p supervisor start toran-proxy-cron" >> /etc/cron.d/toran-proxy
echo "" >> /etc/cron.d/toran-proxy

# Load toran logs
mkdir -p $DATA_DIRECTORY/logs/toran
rm -f $WORK_DIRECTORY/app/logs/prod.log
rm -f $WORK_DIRECTORY/app/logs/downloads.private.log
touch $DATA_DIRECTORY/logs/toran/prod.log
touch $DATA_DIRECTORY/logs/toran/downloads.private.log
ln -s $DATA_DIRECTORY/logs/toran/prod.log $WORK_DIRECTORY/app/logs/prod.log
ln -s $DATA_DIRECTORY/logs/toran/downloads.private.log $WORK_DIRECTORY/app/logs/downloads.private.log

# Loading Toran files permissions for running
echo "Loading permissions..."
chmod -R 777 $WORK_DIRECTORY/app/cache
chown -R www-data:www-data $WORK_DIRECTORY $DATA_DIRECTORY/packagist $DATA_DIRECTORY/toran $DATA_DIRECTORY/mirrors

# Installing ssh keys
if [ -e "$DATA_DIRECTORY/ssh" ]; then
    echo "Installing ssh config..."
    mkdir /var/www/.ssh

    if [ -e "$DATA_DIRECTORY/ssh/id_rsa" ]; then
        chgrp 545 $DATA_DIRECTORY/ssh/id_rsa
        chmod 600 $DATA_DIRECTORY/ssh/id_rsa
        ln -s $DATA_DIRECTORY/ssh/id_rsa /var/www/.ssh/id_rsa
    fi

    if [ -e "$DATA_DIRECTORY/ssh/id_rsa.pub" ]; then
        chmod 644 $DATA_DIRECTORY/ssh/id_rsa.pub
        ln -s $DATA_DIRECTORY/ssh/id_rsa.pub /var/www/.ssh/id_rsa.pub
    fi

    if [ -e "$DATA_DIRECTORY/ssh/known_hosts" ]; then
        chmod 644 $DATA_DIRECTORY/ssh/known_hosts
        ln -s $DATA_DIRECTORY/ssh/known_hosts /var/www/.ssh/known_hosts
    fi

    if [ -e "$DATA_DIRECTORY/ssh/config" ]; then
        chmod 644 $DATA_DIRECTORY/ssh/config
        ln -s $DATA_DIRECTORY/ssh/config /var/www/.ssh/config
    fi

    chmod 700 $DATA_DIRECTORY/ssh
    chown -R www-data:www-data $DATA_DIRECTORY/ssh
fi

# Configure PHP
echo "Configure PHP..."
PHP_TIMEZONE=${PHP_TIMEZONE:-Asia/Shanghai}
sed -i "s|;date.timezone =.*|date.timezone = ${PHP_TIMEZONE}|g" /etc/php/7.1/fpm/php.ini
sed -i "s|;date.timezone =.*|date.timezone = ${PHP_TIMEZONE}|g" /etc/php/7.1/cli/php.ini
mkdir -p $DATA_DIRECTORY/logs/php-fpm
mkdir -p $DATA_DIRECTORY/logs/php-cli
sed -i "s|;error_log = php_errors.log|error_log = ${DATA_DIRECTORY}/logs/php-fpm/errors.log|g" /etc/php/7.1/fpm/php.ini
sed -i "s|;error_log = php_errors.log|error_log = ${DATA_DIRECTORY}/logs/php-cli/errors.log|g" /etc/php/7.1/cli/php.ini

#Configure Nginx
echo "Detecting HTTP Basic Authentication Configuration"
if [ "${TORAN_AUTH_ENABLE}" != "false" ]; then
    if [ ! -e /etc/nginx/.htpasswd ]; then
        echo "Generating .htpasswd file"
        htpasswd -bc /etc/nginx/.htpasswd ${TORAN_AUTH_USER} ${TORAN_AUTH_PASSWORD}
    else
        echo "Skipping .htpasswd generation - already exists."
    fi

    echo "Configuring Nginx for HTTP Basic Authentication..."
    sed -i "s|# auth_basic|auth_basic|g" /etc/nginx/sites-available/toran-proxy-http.conf
    sed -i "s|# auth_basic|auth_basic|g" /etc/nginx/sites-available/toran-proxy-https-reverse.conf
    sed -i "s|# auth_basic|auth_basic|g" /etc/nginx/sites-available/toran-proxy-https.conf
fi
echo "Loading Nginx vhosts..."
rm -f /etc/nginx/sites-enabled/*
if [ "${TORAN_HTTPS}" == "true" ]; then
    if [ "${TORAN_REVERSE}" == "false" ]; then
        echo "Loading HTTPS Certificates..."
        if [ ! -e "${DATA_DIRECTORY}/certs/toran-proxy.key" ] && [ ! -e "${DATA_DIRECTORY}/certs/toran-proxy.crt" ]; then
            echo "Generating self-signed HTTPS Certificates..."
            mkdir -p ${DATA_DIRECTORY}/certs
            openssl req \
                -x509 \
                -nodes \
                -days 365 \
                -newkey rsa:2048 \
                -keyout "${DATA_DIRECTORY}/certs/toran-proxy.key" \
                -out "${DATA_DIRECTORY}/certs/toran-proxy.crt" \
                -subj "/C=SS/ST=SS/L=SelfSignedCity/O=SelfSignedOrg/CN=${TORAN_HOST}"
        elif [ -e "${DATA_DIRECTORY}/certs/toran-proxy.key" ] && [ -e "${DATA_DIRECTORY}/certs/toran-proxy.crt" ]; then
            echo "Using provided HTTPS Certificates..."
        else
            if [ ! -e "${DATA_DIRECTORY}/certs/toran-proxy.key" ]; then
                echo "ERROR: "
                echo "  File toran-proxy.key exists in folder certs/ but no toran-proxy.crt"
                exit 1
            else
                echo "ERROR: "
                echo "  File toran-proxy.crt exists in folder certs/ but no toran-proxy.key"
                exit 1
            fi
        fi

        # Add certificates trusted(Ubuntu下添加根证书)
        # 只要将证书(扩展名为crt)复制到/usr/local/share/ca-certificates文件夹，然后运行update-ca-certificates即可
        ln -s $DATA_DIRECTORY/certs /usr/local/share/ca-certificates/toran-proxy
        update-ca-certificates

        sed -i "s|TORAN_HTTP_PORT|$TORAN_HTTP_PORT|g" /etc/nginx/sites-available/toran-proxy-https.conf
        sed -i "s|TORAN_HTTPS_PORT|$TORAN_HTTPS_PORT|g" /etc/nginx/sites-available/toran-proxy-https.conf
        ln -s /etc/nginx/sites-available/toran-proxy-https.conf /etc/nginx/sites-enabled/toran-proxy-https.conf
    else
        sed -i "s|TORAN_HTTP_PORT|$TORAN_HTTP_PORT|g" /etc/nginx/sites-available/toran-proxy-https-reverse.conf
        sed -i "s|TORAN_HTTPS_PORT|$TORAN_HTTPS_PORT|g" /etc/nginx/sites-available/toran-proxy-https-reverse.conf
        ln -s /etc/nginx/sites-available/toran-proxy-https-reverse.conf /etc/nginx/sites-enabled/toran-proxy-https-reverse.conf
    fi
else
    sed -i "s|TORAN_HTTP_PORT|$TORAN_HTTP_PORT|g" /etc/nginx/sites-available/toran-proxy-http.conf
    ln -s /etc/nginx/sites-available/toran-proxy-http.conf /etc/nginx/sites-enabled/toran-proxy-http.conf
fi
mkdir -p $DATA_DIRECTORY/logs/nginx # Logs
chown -R www-data:www-data $DATA_DIRECTORY/logs

# Final: Start Toran Proxy service
echo "Starting Toran Proxy..."
/usr/bin/supervisord -c /etc/supervisor/supervisord.conf
