


PREFIX = lx1036
IMAGE = nginx-ssl
TAG = 1.0.0

ca:
	/bin/zsh ./ssl/ca.sh

admin:
	/bin/zsh ./ssl/admin.sh

build:
	docker build --tag ${PREFIX}/${IMAGE}:${TAG} -f Dockerfile .

run:
	docker run -p 80:80 -p 443:443 ${PREFIX}/${IMAGE}:${TAG}
