
.PHONY:	build push

PREFIX = lx1036
IMAGE = elasticsearch
TAG = 7.4.2

build:
	docker build --tag ${PREFIX}/${IMAGE}:${TAG} -f Dockerfile-k8s .
	docker build --tag ${PREFIX}/${IMAGE}:latest -f Dockerfile-k8s .

push:
	docker build --tag ${PREFIX}/${IMAGE}:${TAG} .
	docker build --tag ${PREFIX}/${IMAGE}:latest .
