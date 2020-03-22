＃添加一个新的API

搭建新的种类（您正在注意[最后章节](./gvks.md＃kinds-and-resources)，对吗？)控制器，我们可以使用`kubebuilder create api`：

```
kubebuilder create api --group batch --version v1 --kind CronJob
```

我们第一次为每个组版本调用此命令时，它将创建新组版本的目录。

在这种情况下，[`api/v1/`]（https://sigs.k8s.io/kubebuilder/docs/book/src/cronjob-tutorial/testdata/project/api/v1）
目录已创建，对应于`batch.tutorial.kubebuilder.io / v1`（请记住我们的[`--domain`设置]中的（cronjob-tutorial.md＃scaffolding-out-our-project）开始？）。

它还为我们的`CronJob`类型添加了一个文件，`api/v1/cronjob_types.go`。每次我们以不同的方式调用命令种类，它将添加一个相应的新文件。

让我们看看开箱即用的产品，然后我们可以继续填写。

{{#literatego ./testdata/emptyapi.go}}

现在我们已经了解了基本结构，让我们对其进行填写！
