FROM golang:1.13.7
WORKDIR /go/src/redis_exporter/
RUN echo "Asia/Shanghai" >> /etc/timezone

ADD *.go /go/src/redis_exporter/
ADD ../../../../vendor /go/src/redis_exporter/vendor

ARG GOARCH="amd64"
ARG SHA1="[no-sha]"
ARG TAG="[no-tag]"

RUN BUILD_DATE=$(date +%F-%T) && CGO_ENABLED=0 GOOS=darwin GOARCH=$GOARCH go build -o /redis_exporter \
    -ldflags  "-s -w -extldflags \"-static\" -X main.BuildVersion=$TAG -X main.BuildCommitSha=$SHA1 -X main.BuildDate=$BUILD_DATE" .

EXPOSE     9121
ENTRYPOINT [ "/redis_exporter" ]
