
FROM nginx:1.17.9

COPY nginx.conf /etc/nginx/
COPY conf.d/default.conf /etc/nginx/conf.d/default.conf
COPY ssl/admin.pem /etc/nginx/cert/admin.pem
COPY ssl/admin-key.pem /etc/nginx/cert/admin-key.pem

#EXPOSE 80
