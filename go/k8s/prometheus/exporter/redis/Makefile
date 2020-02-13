.PHONY: dev

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
VERSION := $(shell git describe --always --tags | grep -Eo "[0-9]+\.[0-9]+[\.[0-9]+]*")
COMMIT := $(shell git rev-parse --short HEAD)
PWD := $(shell pwd)
USER := $(shell id -nu)
USERID := $(shell id -u)
GROUP := $(shell id -g)

dev:
	docker-compose -f ./docker-compose.yml up -d

build:
	docker build -f Dockerfile -t lx1036/redis_exporter:COMMIT .

push:
	docker push lx1036/redis_exporter:COMMIT

publish: build push
