
# 如何在Ingress里自定义配置nginx.conf
三种方式: Annotation, ConfigMap, Custom Template


## Annotation


### 如何通过Ingress模板自定义配置nginx.conf
解决方案是通过 [server-snippet annotation](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#server-snippet) 来解决。


## ConfigMap



## Custom Template


