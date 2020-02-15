
# **[What is traefik](https://traefik.io/)**
An open-source reverse proxy and load balance for HTTP and TCP-based applications.


# Traefik Setup
```shell script
wget https://github.com/containous/traefik/releases/download/v2.0.4/traefik_v2.0.4_darwin_amd64.tar.gz
tar -zxf traefik_v2.0.4_darwin_amd64.tar.gz
mv traefik /usr/local/bin/traefik
brew install traefik #mac
traefik --configFile=traefik.sample.toml
```


# How does Traefik discover the services?
https://juejin.im/entry/5b752fbaf265da28140e5bdb
https://docs.traefik.io/v2.1/
