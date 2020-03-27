
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

#Starting a single node cluster with Dockeredit
single-node:
	docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.1

multi-nodes:
	docker-compose -f ./docker-compose.yml up
