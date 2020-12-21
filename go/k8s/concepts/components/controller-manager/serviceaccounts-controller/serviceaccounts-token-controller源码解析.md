

# Kubernetes学习笔记之ServiceAccount TokensController源码解析
本文章基于k8s release-1.17分支代码，代码位于`pkg/controller/serviceaccount`目录，代码：**[tokens_controller.go](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/controller/serviceaccount/tokens_controller.go)** 。

## Overview
在**[Kubernetes学习笔记之ServiceAccount AdmissionController源码解析]()**文章中知道一个ServiceAccount对象都会引用一个
`type="kubernetes.io/service-account-token"` 的secret对象，这个secret对象内的`ca.crt`、`namespace`和`token`数据会被挂载到pod内的
每一个容器，供调用api-server时认证授权使用。

当创建一个ServiceAccount对象时，引用的 `type="kubernetes.io/service-account-token"` 的secret对象会自动创建。比如：
```shell
kubectl create sa test-sa1 -o yaml
kubectl get sa test-sa1 -o yaml
kubectl get secret test-sa1-token-jg6lm -o yaml
```

![serviceaccount_token](./imgs/serviceaccount_token.png)

问题是，这是怎么做到的呢？


## 源码解析

### TokensController实例化
实际上这是由kube-controller-manager的TokenController实现的，该kube-controller-manager进程在启动时参数有`--root-ca-file`和`--service-account-private-key-file`，
其中，`--root-ca-file`就是上图中的`ca.crt`数据，`--service-account-private-key-file`是用来签名上图中的jwt token数据，即`token`字段值。

当kube-controller-manager进程在启动时，会首先实例化TokensController，并传递实例化所需相关参数。
其中，从启动参数中读取ca根证书和私钥文件内容，并且使用`serviceaccount.JWTTokenGenerator()`函数生成jwt token，
代码在 **[L546-L592](https://github.com/kubernetes/kubernetes/blob/release-1.17/cmd/kube-controller-manager/app/controllermanager.go#L546-L592)**：
```go
func (c serviceAccountTokenControllerStarter) startServiceAccountTokenController(ctx ControllerContext) (http.Handler, bool, error) {
	// ...
	// 读取--service-account-private-key-file私钥文件
	privateKey, err := keyutil.PrivateKeyFromFile(ctx.ComponentConfig.SAController.ServiceAccountKeyFile)
	if err != nil {
		return nil, true, fmt.Errorf("error reading key for service account token controller: %v", err)
	}

	// 读取--root-ca-file的值作为ca，没有传则使用kubeconfig文件内的ca值
	var rootCA []byte
	if ctx.ComponentConfig.SAController.RootCAFile != "" {
		if rootCA, err = readCA(ctx.ComponentConfig.SAController.RootCAFile); err != nil {
			return nil, true, fmt.Errorf("error parsing root-ca-file at %s: %v", ctx.ComponentConfig.SAController.RootCAFile, err)
		}
	} else {
		rootCA = c.rootClientBuilder.ConfigOrDie("tokens-controller").CAData
	}

	// 使用tokenGenerator来生成jwt token，并且使用--service-account-private-key-file私钥来签名jwt token
	tokenGenerator, err := serviceaccount.JWTTokenGenerator(serviceaccount.LegacyIssuer, privateKey)
	//...
	
	// 实例化TokensController
	controller, err := serviceaccountcontroller.NewTokensController(
		ctx.InformerFactory.Core().V1().ServiceAccounts(), // ServiceAccount informer
		ctx.InformerFactory.Core().V1().Secrets(), // Secret informer
		c.rootClientBuilder.ClientOrDie("tokens-controller"),
		serviceaccountcontroller.TokensControllerOptions{
			TokenGenerator: tokenGenerator,
			RootCA:         rootCA,
		},
	)
	// ...
	// 消费队列数据
	go controller.Run(int(ctx.ComponentConfig.SAController.ConcurrentSATokenSyncs), ctx.Stop)

	// 启动ServiceAccount informer和Secret informer
	ctx.InformerFactory.Start(ctx.Stop)

	return nil, true, nil
}
```

TokensController实例化时，会去监听ServiceAccount和`kubernetes.io/service-account-token`类型的Secret对象，并设置监听器：
```go
func NewTokensController(serviceAccounts informers.ServiceAccountInformer, secrets informers.SecretInformer, cl clientset.Interface, options TokensControllerOptions) (*TokensController, error) {
    e := &TokensController{
        // ...
    	// 分别为service和secret创建对应的限速队列queue，用来存储事件数据
        syncServiceAccountQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "serviceaccount_tokens_service"),
        syncSecretQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "serviceaccount_tokens_secret"),
    }
	// ...
	e.serviceAccounts = serviceAccounts.Lister()
	e.serviceAccountSynced = serviceAccounts.Informer().HasSynced
	serviceAccounts.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    e.queueServiceAccountSync,
			UpdateFunc: e.queueServiceAccountUpdateSync,
			DeleteFunc: e.queueServiceAccountSync,
		},
		options.ServiceAccountResync,
	)

	// ...
	secrets.Informer().AddEventHandlerWithResyncPeriod(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Secret:
					return t.Type == v1.SecretTypeServiceAccountToken // 这里过滤出"kubernetes.io/service-account-token"类型的secret
				default:
					utilruntime.HandleError(fmt.Errorf("object passed to %T that is not expected: %T", e, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    e.queueSecretSync,
				UpdateFunc: e.queueSecretUpdateSync,
				DeleteFunc: e.queueSecretSync,
			},
		},
		options.SecretResync,
	)

	return e, nil
}
// 把service对象存进syncServiceAccountQueue
func (e *TokensController) queueServiceAccountSync(obj interface{}) {
    if serviceAccount, ok := obj.(*v1.ServiceAccount); ok {
        e.syncServiceAccountQueue.Add(makeServiceAccountKey(serviceAccount))
    }
}
// 把secret对象存进syncSecretQueue
func (e *TokensController) queueSecretSync(obj interface{}) {
    if secret, ok := obj.(*v1.Secret); ok {
        e.syncSecretQueue.Add(makeSecretQueueKey(secret))
    }
}
```

把数据存入队列后，goroutine调用controller.Run()来消费队列数据，执行具体业务逻辑：
```go
func (e *TokensController) Run(workers int, stopCh <-chan struct{}) {
	// ...
	for i := 0; i < workers; i++ {
		go wait.Until(e.syncServiceAccount, 0, stopCh)
		go wait.Until(e.syncSecret, 0, stopCh)
	}
	<-stopCh
	// ...
}
```

### Controller业务逻辑

#### ServiceAccount的增删改查
当用户增删改查ServiceAccount时，需要判断两个业务逻辑：当删除ServiceAccount时，需要删除其引用的Secret对象；当添加/更新ServiceAccount时，
需要确保引用的Secret对象存在，如果不存在，则创建个新Secret对象。可见代码：
```go
func (e *TokensController) syncServiceAccount() {
	// ...
	sa, err := e.getServiceAccount(saInfo.namespace, saInfo.name, saInfo.uid, false)
	switch {
	case err != nil:
		klog.Error(err)
		retry = true
	case sa == nil:
		// 该service account已经被删除，需要删除其引用的secret对象
		sa = &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: saInfo.namespace, Name: saInfo.name, UID: saInfo.uid}}
		retry, err = e.deleteTokens(sa)
	default:
		// 创建/更新service account时，需要确保其引用的secret对象存在，不存在则新建一个secret对象
		retry, err = e.ensureReferencedToken(sa)
		// ...
	}
}
```

先看如何删除其引用的secret对象的业务逻辑：



再看如何新建secret对象的业务逻辑：



#### Secret的增删改查



