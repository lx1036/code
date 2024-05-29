
CONTAINER_ENGINE := docker
GO_VERSION := 1.20.5
CURDIR := $(shell pwd)

GO_CONTAINER := $(CONTAINER_ENGINE) run --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) golang:$(GO_VERSION)
GOIMPORTS_VERSION ?= v0.1.12

SWAGGER_VERSION := v0.25.0
SWAGGER := $(CONTAINER_ENGINE) run -u $(shell id -u):$(shell id -g) --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) --entrypoint swagger quay.io/goswagger/swagger:$(SWAGGER_VERSION)

generate-api: api/v1/openapi.yaml ## Generate cilium-agent client, model and server code from openapi spec.
	@$(ECHO_GEN)api/v1/openapi.yaml
	-$(QUIET)$(SWAGGER) generate server -s server -a restapi \
		-t api/v1 \
		-f api/v1/openapi.yaml \
		--default-scheme=unix \
		-C api/v1/cilium-server.yml \
		-r hack/spdx-copyright-header.txt
	-$(QUIET)$(SWAGGER) generate client -a restapi \
		-t api/v1 \
		-f api/v1/openapi.yaml
#		-r hack/spdx-copyright-header.txt
	@# sort goimports automatically
	-$(QUIET)$(GO_CONTAINER) bash -c "go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) && ./contrib/scripts/format-api.sh api/v1/client/ api/v1/models/ api/v1/server/"
