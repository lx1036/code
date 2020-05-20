

.PHONY:	build

PREFIX = lx1036
IMAGE = golang-metrics
TAG = v1.0.0

build:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -v -o ./bin/app ./main.go
	#docker build --tag ${PREFIX}/${IMAGE}:${TAG} -f Dockerfile .
