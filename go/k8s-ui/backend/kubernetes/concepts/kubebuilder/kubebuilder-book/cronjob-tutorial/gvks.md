＃ Groups and Versions and Kinds, oh my!

实际上，在开始使用API​​之前，我们应该先谈谈术语一点点。

在Kubernetes中谈论API时，我们经常使用4个术语：*groups*，*versions*，*kinds* 和 *resources*。

## Groups and Versions

Kubernetes中的 *API Group* 只是相关的集合功能。每个组都有一个或多个 *version*，以名称命名建议，让我们随着时间的推移更改API的工作方式。

## Kinds and Resources

每个API组版本都包含一种或多种API类型，我们称之为 *Kinds*。虽然种类可能会在版本之间更改表格，但每种表格都必须能够以某种方式存储其他形式的所有数据（我们可以存储字段或注释中的数据）。这意味着使用较旧的API版本不会导致新数据丢失或损坏。看到
[Kubernetes API准则]（https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md）了解更多信息。

您还会偶尔听到有关 *resources* 的提及。资源就是在API中使用种类。通常，两者之间存在一对一的映射种类和资源。例如，“ pods”资源对应于`Pod`种。
但是，有时相同的种类可能会由多个资源。例如，“Scale” 种类由所有比例返回子资源，例如“部署/规模”或“复制品/规模”。
这是是什么让Kubernetes Horizo​​ntalPodAutoscaler与不同的资源。但是，对于CRD，每种类型将对应于单一资源。

请注意，资源始终是小写的，按照惯例，Kind 的小写形式。

## 那么，这与Go相对应？

当我们在特定的小组版本中提及一种类型时，我们将其称为 *GroupVersionKind*，或简称GVK。与资源和GVR相同。如
我们很快就会看到，每个GVK对应于一个给定的根Go类型一套。

现在我们有了明确的术语，我们可以*实际上*创建我们的API！

## 但是Scheme到底是什么？

我们之前看到的“Scheme”只是一种跟踪Go类型的方法对应于给定的GVK（不要被它的文档搞糊涂了 [godocs](https://godoc.org/k8s.io/apimachinery/pkg/runtime#Scheme)）。

例如，假设我们将 `“tutorial.kubebuilder.io/api/v1”.CronJob{}` 类型位于 `batch.tutorial.kubebuilder.io/v1` API组（暗指它具有种“ CronJob”）。

然后，我们可以在给定JSON的基础上构造一个新的＆CronJob {} API服务器说

```json
{
    "kind": "CronJob",
    "apiVersion": "batch.tutorial.kubebuilder.io/v1",
    ...
}
```

或在我们提交“＆CronJob{}”时正确查找组版本在更新中。
