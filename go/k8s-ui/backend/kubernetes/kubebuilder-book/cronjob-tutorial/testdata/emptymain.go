/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// +kubebuilder:docs-gen:collapse=Apache License

/*

我们的程序包从一些基本的进口开始。 尤其：

-核心库[controller-runtime]（https://godoc.org/sigs.k8s.io/controller-runtime）
-默认的控制器运行时日志记录Zap（稍后会详细介绍）

*/

package main

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

/*
每套控制器都需要一个[*Scheme*](https://book.kubebuilder.io/cronjob-tutorial/gvks.html#err-but-whats-that-scheme-thing)，
提供了Kinds及其对应的Go类型之间的映射。 好在编写API定义时，再多谈谈Kinds，所以为了以后请先记住。
*/
var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {

	// +kubebuilder:scaffold:scheme
}

/*
至此，我们的主要功能非常简单：
- 我们为指标设置了一些基本标志。

- 我们实例化一个 [*manager*](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/manager#Manager)，跟踪运行我们所有的控制器以及设置共享缓存和客户端到API服务器（请注意，我们告诉经理有关我们的计划）。

- 我们运行 manager，而经理又运行所有控制器和网络挂钩。管理器设置为运行，直到收到正常关闭信号为止。这样，当我们在Kubernetes上运行时，我们会表现得很优雅吊舱终止。

虽然我们还没有什么可以运行的，但请记住 `+ kubebuilder:scaffold:builder`注释是-那里的事情会变得有趣不久。

*/

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme, MetricsBindAddress: metricsAddr})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	/*
		Note that the Manager can restrict the namespace that all controllers will watch for resources by:
	*/

	mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
	})

	/*
		上面的示例将项目范围更改为单个命名空间。 在这种情况下，还建议通过替换默认名称来将提供的授权限制为此名称空间
		将ClusterRole和ClusterRoleBinding分别设置为Role和RoleBinding。
		有关更多信息，请参阅有关使用[RBAC授权]的kubernetes文档（https://kubernetes.io/docs/reference/access-authn-authz/rbac/）。

		同样，可以使用 MultiNamespacedCacheBuilder 监视一组特定的名称空间：
	*/

	var namespaces []string // List of Namespaces

	mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		NewCache:           cache.MultiNamespacedCacheBuilder(namespaces),
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})

	/*
		更多信息请查看 [MultiNamespacedCacheBuilder](https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder)
	*/

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
