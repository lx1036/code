
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

# 执行后，就可以在Kibana里看到每一个容器的日志：
# filebeat去/var/lib/docker/containers/*/*.log里收集日志(每一个log文件都会启动一个日志收割机harvester)，发给elasticsearch，然后在kibana里展示。
multi-nodes:
	docker-compose -f ./docker-compose.yml -p elk down && docker-compose -f ./docker-compose.yml -p elk up --remove-orphans -d
