# Docs
https://github.com/bakins/php-fpm-exporter
https://easyengine.io/tutorials/php/fpm-status-page/

# Test
```shell script
php-fpm --nodaemonize --fpm-config ./php-fpm.conf
nginx -c pwd/nginx.conf
php-fpm-exporter
```
