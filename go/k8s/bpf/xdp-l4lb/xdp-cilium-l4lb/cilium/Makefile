
# Makefile.quiet 里可以使得 `generate-api: api/v1/openapi.yaml` 而没有 `api/v1/openapi.yaml` 定义
include Makefile.quiet

include Makefile.defs

#-include Makefile.docker

CONTAINER_ENGINE := docker
GO_VERSION := 1.20.5
CURDIR := $(shell pwd)

GO_CONTAINER := $(CONTAINER_ENGINE) run --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) golang:$(GO_VERSION)
GOIMPORTS_VERSION ?= v0.1.12

# https://goswagger.io/go-swagger/
SWAGGER_VERSION := v0.25.0
SWAGGER := $(CONTAINER_ENGINE) run -u $(shell id -u):$(shell id -g) --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) --entrypoint swagger quay.io/goswagger/swagger:$(SWAGGER_VERSION)
SWAGGER-CLI := go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.5


generate-api: api/v1/openapi.yaml ## Generate cilium-agent client, model and server code from openapi spec.
	$(ECHO_GEN)api/v1/openapi.yaml
	swagger generate server -s server -a restapi \
		-t api/v1 \
		-f api/v1/openapi.yaml \
		--default-scheme=unix \
		-C api/v1/cilium-server.yml

	swagger generate client -a restapi \
		-t api/v1 \
		-f api/v1/openapi.yaml

	@# sort goimports automatically
	#$(GO_CONTAINER) bash -c "go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) && ./contrib/scripts/format-api.sh api/v1/client/ api/v1/models/ api/v1/server/"
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) && ./contrib/scripts/format-api.sh api/v1/client/ api/v1/models/ api/v1/server/
