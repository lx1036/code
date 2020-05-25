FROM openresty/openresty:1.15.8.3-1-centos
LABEL maintainer="lx1036@126.com"
RUN echo "Asia/Shanghai" >> /etc/timezone
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY ./nginx.conf /usr/local/openresty/nginx/conf/nginx.conf

# lua
COPY ./prometheus/ /var/www/html/prometheus/
COPY ./pc/ /var/www/html/pc/
COPY ./mobile/ /var/www/html/mobile/

ENTRYPOINT /usr/local/openresty/nginx/sbin/nginx -c /usr/local/openresty/nginx/conf/nginx.conf
