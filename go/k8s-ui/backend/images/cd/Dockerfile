FROM golang:1.14.0

RUN echo "Asia/Shanghai" >> /etc/timezone
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENV GOPROXY https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct

LABEL maintainer="lx1036@126.com"

COPY ./src /go/src
WORKDIR /go/src/app
RUN go build -mod=vendor -v -o /cd app
RUN rm -fr /go/src

ENTRYPOINT /app
