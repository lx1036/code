
# **[What is traefik](https://traefik.io/)**
An open-source reverse proxy and load balancer for HTTP and TCP-based applications.


# Traefik Setup
```shell script
wget https://github.com/containous/traefik/releases/download/v2.0.4/traefik_v2.0.4_darwin_amd64.tar.gz
tar -zxf traefik_v2.0.4_darwin_amd64.tar.gz
mv traefik /usr/local/bin/traefik
traefik --configFile=traefik.sample.toml
```
