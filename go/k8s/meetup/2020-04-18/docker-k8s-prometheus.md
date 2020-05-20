
# Docker

```shell script
# VM
sudo apt update && sudo apt install -y nginx

# Docker
sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update && sudo apt-get install -y docker-ce docker-ce-cli containerd.io

sudo docker run -p 8088:80 --name nginx-demo -d nginx
sudo docker run -p 8089:80 --name nginx-demo2 -v /home/lx1036/index2.html:/usr/share/nginx/html/index.html -d nginx
sudo docker container exec -it nginx-demo /bin/bash

# Dockerfile
make build
make push

```



# Kubernetes






# Prometheus




