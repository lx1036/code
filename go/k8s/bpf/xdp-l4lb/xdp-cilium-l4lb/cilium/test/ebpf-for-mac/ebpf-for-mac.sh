

# mac 本地运行，需要 volume debugfs(验证可用)
# docker volume create --driver local --opt type=debugfs --opt device=debugfs debugfs
docker stop ebpf-for-mac && docker rm ebpf-for-mac
docker run -it --name ebpf-for-mac --privileged -v debugfs:/sys/kernel/debug:ro \
-v /lib/modules:/lib/modules:ro -v /etc/localtime:/etc/localtime:ro --pid=host \
-v /Users/liuxiang/Code/code:/mnt/code \
-v /Users/liuxiang/go/pkg/mod:/root/go/pkg/mod \
lx1036/ebpf-for-mac:2.2 /bin/bash


