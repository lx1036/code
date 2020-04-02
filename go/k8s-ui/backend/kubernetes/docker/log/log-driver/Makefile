
PREFIX = lx1036
IMAGE = log-driver
TAG = 1.0.0


all: clean rootfs create enable ls install clean

clean:
	@echo "rm -rf ./plugin"
	@sudo rm -rf ./plugin

rootfs:
	echo "docker build: rootfs image with docker-volume-sshfs"
	docker build -q -t ${PREFIX}/${IMAGE}:rootfs .
	echo "create rootfs directory in ./plugin/rootfs"
	mkdir -p ./plugin/rootfs
	docker create --name tmp ${PREFIX}/${IMAGE}:rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	echo "copy config.json to ./plugin/"
	cp config.json ./plugin/
	docker rm -vf tmp

create:
	@echo "remove existing plugin ${PREFIX}/${IMAGE}:${TAG} if exists"
	docker plugin rm -f ${PREFIX}/${IMAGE}:${TAG} || true
	echo "create new plugin ${PREFIX}/${IMAGE}:${TAG} from ./plugin"
	docker plugin create ${PREFIX}/${IMAGE}:${TAG} ./plugin

enable:
	echo "enable plugin ${PREFIX}/${IMAGE}:${TAG}"
	docker plugin enable ${PREFIX}/${IMAGE}:${TAG}

install:
	docker plugin install ${PREFIX}/${IMAGE}:${TAG}

ls:
	docker plugin ls
