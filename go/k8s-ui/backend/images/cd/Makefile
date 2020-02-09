.PHONY: build

PWD := $(shell pwd)
USER := $(shell id -u)
GROUP := $(shell id -g)
VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1 --no-walk))


build:
	docker build -f Dockerfile -t continue_deployment:$(VERSION) .
