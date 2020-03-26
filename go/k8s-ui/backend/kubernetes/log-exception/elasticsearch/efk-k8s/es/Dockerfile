
FROM golang:1.14.0 as builder
ENV GOPROXY https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-w" -o /elasticsearch_logging_discovery /elasticsearch_logging_discovery.go

#https://www.docker.elastic.co/#
#https://www.elastic.co/guide/en/elasticsearch/reference/7.6/docker.html
FROM docker.elastic.co/elasticsearch/elasticsearch:7.6.1
VOLUME ["/data"]
EXPOSE 9200 9300
COPY --from=builder /elasticsearch_logging_discovery /bin
COPY run.sh bin/
COPY config/elasticsearch.yml config/log4j2.properties config/

USER root
RUN chown -R elasticsearch:elasticsearch ./
CMD ["bin/run.sh"]
