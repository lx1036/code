
USER := $(shell id -u)
GROUP := $(shell id -g)

pc:
	docker container rm openresty-pc
	docker run -p 8088:80 --name openresty-pc \
	-u $(USER):$(GROUP) \
	-v $(PWD)/index.html:/var/www/html/pc/index.html \
	-v $(PWD)/openresty-default.conf:/etc/openresty/nginx.conf \
	openresty/openresty
