＃基本项目中有什么？

在搭建新项目时，Kubebuilder为我们提供了一些基本样板。

##建立基础架构

首先，构建项目的基本基础架构：

<details> <summary>`go.mod`：与我们的项目匹配的新Go模块，带有基本依赖项</summary>

```go
{{#include ./testdata/project/go.mod}}
```
</details>

<details> <summary>`Makefile`：设置用于构建和部署控制器的目标</summary>

```makefile
{{#include ./testdata/project/Makefile}}
```
</details>

<details> <summary>`PROJECT`：用于搭建新组件的Kubebuilder元数据</summary>

```yaml
{{## ./testdata/project/PROJECT}}
```
</details>

##启动配置

我们还会在[`config /`](https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project/config)
目录。现在，它只包含[Kustomize](https://sigs.k8s.io/kustomize)所需的YAML定义在集群上启动我们的控制器，但是一旦我们开始编写我们的控制器，它还将包含我们的CustomResourceDefinitions，RBAC配置和WebhookConfigurations。

[`config/default`](https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project/config/default)包含[Kustomize基础](https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/testdata/project/config/default/kustomization.yaml)启动
标准配置的控制器。

每个其他目录包含不同的配置，重构成自己的基础：

-[`config/manager`](https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project/config/manager)：以以下方式启动控制器中的豆荚簇

-[`config/rbac`](https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project/config/rbac)：运行所需的权限您的他们自己的服务帐户下的控制者

##入口点

最后，但同样重要的是，Kubebuilder搭建了基本我们项目的入口点：“ main.go”。让我们看看下一个...
