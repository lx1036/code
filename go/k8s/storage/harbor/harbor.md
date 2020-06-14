
# Harbor Components
**[Architecture Overview of Harbor](https://github.com/goharbor/harbor/wiki/Architecture-Overview-of-Harbor)**



|Component|	Version|
|---|---|
|Postgresql	|9.6.10-1.ph2|
|Redis	|4.0.10-1.ph2|
|Clair	|2.0.8|
|Beego	|1.9.0|
|Chartmuseum|	0.9.0|
|Docker/distribution|	2.7.1|
|Docker/notary|	0.6.1|
|Helm|	2.9.1|
|Swagger-ui	|3.22.1|

* Clair 是coreos 开源的容器漏洞扫描工具。
* Redis 是
* harbor-adminserver: 是harbor系统管理接口，可以修改系统配置以及获取系统信息。
* harbor-jobservice: 是harbor的job管理模块，job在harbor里面主要是为了镜像仓库之间同步使用的。
* harbor-db:是harbor的数据库(MySQL)，这里保存了系统的job以及项目、人员权限管理。由于本harbor的认证也是通过数据，在生产环节大多对接到企业的ldap中。
* harbor-ui:是web管理页面，主要是前端的页面和后端CURD的接口。
* nginx:负责流量转发和安全验证，对外提供的流量都是从nginx中转，所以开放https的443端口，它将流量分发到后端的ui和正在Docker镜像存储的docker registry。
* registry:由Docker官方的开源registry 镜像构成的容器实例。



# Docs
**[使用Harbor搭建企业级的Docker私有镜像库](https://www.jianshu.com/p/95191c4eed92)**


## harbor-ui 和 harbor-registry 交互


