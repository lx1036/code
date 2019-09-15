#!/usr/bin/env bash

docker build -t dockerhosts ./dockerhosts/
docker rm -f dockerhosts1 dockerhosts2
docker run --rm -d -p 4422:22 -p 8801:80 --name dockerhosts1 dockerhosts:latest
docker run --rm -d -p 4423:22 -p 8802:80 --name dockerhosts2 dockerhosts:latest
ssh-copy-id root@localhost -p 4422
ssh-copy-id root@localhost -p 4423

ansible-playbook ./index.yml -i ./hosts --ask-pass
