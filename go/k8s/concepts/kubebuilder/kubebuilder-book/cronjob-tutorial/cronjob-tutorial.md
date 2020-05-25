
> **[原文链接](https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/cronjob-tutorial.md)**

＃教程：构建CronJob

太多的教程以一些真正的设计或玩具开始了解基础知识，然后在更多方面停滞不前的应用程序复杂的东西。相反，本教程将带您（几乎）完成所有操作使用 Kubebuilder 的全部复杂性，从简单建立功能齐全的产品。

让我们假装（当然，这有点虚构）终于厌倦了非 Kubebuilder 的维护负担在Kubernetes中实现CronJob控制器，我们希望使用KubeBuilder重写它。

*CronJob* 控制器的工作（无双关）是一次性运行Kubernetes集群上的任务定期进行。它是通过建立在* Job *控制器之上，该控制器的任务是运行一次性任务一次，看到他们完成。

除了尝试重新编写Job控制器外，我们还将借此机会了解如何与外部类型进行交互。

<aside class =“note”>

<h1>跟进vs向前跳</h1>

请注意，本教程的大部分内容是从具有读写能力的Go文件生成的，存放在书的源目录中：[docs/book/src/cronjob-tutorial/testdata] [tutorial-source]。饱满的可运行的项目位于[project] [tutorial-project-source]中，而
中间文件直接位于[testdata] [tutorial-source]下目录。

[tutorial-source]：https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata

[tutorial-project-source]：https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project

</aside>

##搭建我们的项目

如[快速入门]（../ quick-start.md）所述，我们需要脚手架出一个新项目。确保您已[安装Kubebuilder]（../ quick-start.md＃installation），然后搭建一个新的项目：

```
＃我们将使用tutorial.kubebuilder.io的域，
＃因此所有API组均为<group> .tutorial.kubebuilder.io。
kubebuilder初始化--domain tutorial.kubebuilder.io
```

现在我们已经有了一个项目，让我们看一下到目前为止，Kubebuilder已经为我们搭建了脚手架。
