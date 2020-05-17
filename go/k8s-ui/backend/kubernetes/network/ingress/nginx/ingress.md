
**[通过traceID实例讲解 Nginx Ingress 参数配置](https://mp.weixin.qq.com/s/tF0A04ZzKzy32c0nu9DUqA)**
**[同一kubernetes部署多个Nginx Ingress Controller](https://mp.weixin.qq.com/s/603OMSae70tNVM27iSV4sQ)**

# Installation

## minikube
```shell script
minikube addons enable/disable ingress
kubectl get pods -n kube-system # nginx-ingress-controller-*
```



# 如何在Ingress里自定义配置nginx.conf
三种方式: Annotation, ConfigMap, Custom Template


## Annotation

#### canary rules(灰色发布)
匹配次序：canary-by-header -> canary-by-cookie -> canary-weight
* nginx.ingress.kubernetes.io/canary: 
* nginx.ingress.kubernetes.io/canary-by-header:
* nginx.ingress.kubernetes.io/canary-by-header-value:
* nginx.ingress.kubernetes.io/canary-by-header-pattern:
* nginx.ingress.kubernetes.io/canary-by-cookie:
* nginx.ingress.kubernetes.io/canary-weight:

#### rewrite rule(重定向)
* nginx.ingress.kubernetes.io/rewrite-target:
* nginx.ingress.kubernetes.io/app-root:


#### session affinity


#### authentication



### 如何通过Ingress模板自定义配置nginx.conf
解决方案是通过 [server-snippet annotation](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#server-snippet) 来解决。


## ConfigMap



## Custom Template


# Log
nginx log 格式如下：
```markdown
log_format upstreaminfo '$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] [$proxy_alternative_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id';
```


# Sticky sessions
**[Sticky sessions](https://github.com/kubernetes/ingress-nginx/blob/master/docs/examples/affinity/cookie/README.md)**

# Install Grafana/Prometheus for metrics
**[monitor](https://kubernetes.github.io/ingress-nginx/user-guide/monitoring/)**


# Nginx Ingress Controller
