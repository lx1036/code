

# Kong
**Kong is an API gateway built on top of Nginx.**

# Install
```shell script
brew tap kong/kong
brew install kong
```

# Arch
Kong主要有三个组件：
```markdown
Kong Server：基于nginx的服务器，用来接收API请求。
Apache Cassandra/PostgreSQL：用来存储操作数据。
Kong Dashboard：官方推荐UI管理工具，也可以使用 restful 方式 管理 admin api。
```

Kong基本概念：

![kong-basic-concepts](./kong-basic-concepts.png)
```markdown
客户端：指下游客户向Kong的代理端口发出请求。
服务：服务实体，是对自己的每个上游服务的抽象。客户请求被转发到该服务。
路由：路由是进入Kong的入口点，并为要匹配的请求定义规则，并路由到给定的Service。服务和路由之间的关系是一对多的关系。
插件：它是在代理生命周期中运行的业务逻辑。可以通过ADMIN API配置插件 - 全局（所有传入流量）或特定的路由和服务。
用户：是调用API 服务时身份认证的凭据
```

**[Kong 网关使用入门](https://juejin.im/post/5d09c307e51d4510a73280c4)**
