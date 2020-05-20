FROM golang:1.14.0

RUN echo "Asia/Shanghai" >> /etc/timezone
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENV GOPROXY https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct

LABEL maintainer="lx1036@126.com"

COPY ./bin/app /var/www/app
ENTRYPOINT ["/var/www/app"]
#CMD ["--url=https://oapi.dingtalk.com/robot/send?access_token=xxx"]

#COPY . /go/src/dingtalk-proxy/
#WORKDIR /go/src/dingtalk-proxy/
#RUN go build -mod=vendor -v -o /go/bin/app ./webhook.go
#CMD ["/go/bin/app"]
