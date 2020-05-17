
PROJECT := lx1036/ubuntu
VERSION := 1.0.1

build:
	docker build -f Dockerfile -t $(PROJECT):$(VERSION) .
	docker push $(PROJECT):$(VERSION)
