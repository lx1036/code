FROM golang:1.14.0

RUN echo "Asia/Shanghai" >> /etc/timezone
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV GOPROXY https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct
LABEL maintainer="lx1036@126.com"

COPY . /var/www/tmp/
WORKDIR /var/www/tmp/

RUN go mod tidy && go mod vendor
RUN go build --ldflags '-extldflags "-static"' -mod=vendor -v -o /usr/bin/docker-log-driver ./main.go
RUN rm -rf /var/www/tmp/

#ENTRYPOINT /usr/bin/docker-log-driver
CMD ["/usr/bin/docker-log-driver"]
