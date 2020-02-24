
IMAGE := lx1036/code-generator
VERSION := 1.0.0
PROJECT_PACKAGE := k8s-lx1036/k8s-ui/backend/kubernetes/crd/code-generator
CODE_GENERATOR_PATH := /go/src/k8s.io/code-generator/

build:
	docker build -f ./Dockerfile -t $(IMAGE):$(VERSION) --build-arg CODE_GENERATOR_PATH=$(CODE_GENERATOR_PATH) ./hack

run:
	docker run -it --rm \
		-v ${PWD}:/go/src/$(PROJECT_PACKAGE) \
		-e CODE_GENERATOR_PATH=$(CODE_GENERATOR_PATH)
		-e PROJECT_PACKAGE=$(PROJECT_PACKAGE) \
		-e CLIENT_GENERATOR_OUT=$(PROJECT_PACKAGE) \
		-e APIS_ROOT=$(PROJECT_PACKAGE) \
		-e GROUPS_VERSION="group1:v1" \
		-e GENERATION_TARGETS="deepcopy,client" \
		lx1036/code-generator:latest
