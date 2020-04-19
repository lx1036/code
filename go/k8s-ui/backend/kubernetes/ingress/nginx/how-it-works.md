
# 工作原理
**[Nginx Ingress Controller 工作原理 中文](https://mp.weixin.qq.com/s/mWf3pZMwe8JXjoE3x4gNEw)**

**[Nginx Ingress Controller 工作原理](https://kubernetes.github.io/ingress-nginx/how-it-works/)**:

![how-it-works](./how-it-works.png)



## nginx.conf
nginx-ingress-controller 主要是拼装 nginx.conf 配置文件，使用 **[lua-nginx-module](https://github.com/openresty/lua-nginx-module)** 模块
来实现，除了 `upstream` 模块以外任何模块发生修改都会 reload nginx。
ingress-nginx-raw.conf 是由 **[nginx template](https://github.com/kubernetes/ingress-nginx/blob/master/rootfs/etc/nginx/template/nginx.tmpl)** 生成的。

## **[Nginx Model](https://kubernetes.github.io/ingress-nginx/how-it-works/#building-the-nginx-model)**


