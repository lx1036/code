

# Kubernetes学习笔记之ServiceAccount TokensController源码解析
本文章基于k8s release-1.17分支代码，代码位于`pkg/controller/serviceaccount`目录，代码：**[tokens_controller.go](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/controller/serviceaccount/tokens_controller.go)** 。

## Overview
在**[Kubernetes学习笔记之ServiceAccount AdmissionController源码解析]()**文章中，知道一个ServiceAccount对象都会引用一个
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
实际上这是由kube-controller-manager的TokenController实现的，kube-controller-manager进程的启动参数有`--root-ca-file`和`--service-account-private-key-file`，
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
	// 
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

先看如何删除其引用的secret对象的业务逻辑，删除逻辑也很简单：

```go
func (e *TokensController) deleteTokens(serviceAccount *v1.ServiceAccount) ( /*retry*/ bool, error) {
	// list出该serviceAccount所引用的所有secret
	tokens, err := e.listTokenSecrets(serviceAccount)
	// ...
	for _, token := range tokens {
		// 在一个个删除secret对象
		r, err := e.deleteToken(token.Namespace, token.Name, token.UID)
		// ...
	}
	// ...
}
func (e *TokensController) deleteToken(ns, name string, uid types.UID) ( /*retry*/ bool, error) {
    // ...
	// 对api-server发起删除secret对象资源的请求
    err := e.client.CoreV1().Secrets(ns).Delete(name, opts)
    // ...
}
```

这里关键是如何找到serviceAccount所引用的所有secret对象，不能通过serviceAccount.secrets字段来查找，因为这个字段值只是所有secrets的部分值。
实际上，从缓存中，首先list出该serviceAccount对象所在的namespace下所有secrets，然后过滤出type=kubernetes.io/service-account-token类型的
secret，然后查找secret annotation中的`kubernetes.io/service-account.name`应该是serviceAccount.Name值，和`kubernetes.io/service-account.uid`
应该是serviceAccount.UID值。只有满足以上条件，才是该serviceAccount所引用的secrets。
首先从缓存中找出该namespace下所有secrets，这里需要注意的是缓存对象updatedSecrets使用的是LRU(Least Recently Used) Cache最少使用缓存，提高查找效率：
```go
func (e *TokensController) listTokenSecrets(serviceAccount *v1.ServiceAccount) ([]*v1.Secret, error) {
	// 从LRU cache中查找出该namespace下所有secrets
	namespaceSecrets, err := e.updatedSecrets.ByIndex("namespace", serviceAccount.Namespace)
	// ...
	items := []*v1.Secret{}
	for _, obj := range namespaceSecrets {
		secret := obj.(*v1.Secret)
		// 判断只有符合相应条件才是该serviceAccount所引用的secret
		if serviceaccount.IsServiceAccountToken(secret, serviceAccount) {
			items = append(items, secret)
		}
	}
	return items, nil
}
// 判断条件
func IsServiceAccountToken(secret *v1.Secret, sa *v1.ServiceAccount) bool {
    if secret.Type != v1.SecretTypeServiceAccountToken {
        return false
    }
    name := secret.Annotations[v1.ServiceAccountNameKey]
    uid := secret.Annotations[v1.ServiceAccountUIDKey]
    if name != sa.Name {
        return false
    }
    if len(uid) > 0 && uid != string(sa.UID) {
        return false
    }
    return true
}
```

所以，当ServiceAccount对象删除时，需要删除其所引用的所有Secrets对象。


再看如何新建secret对象的业务逻辑。当新建或更新ServiceAccount对象时，需要确保其引用的Secrets对象存在，不存在就需要新建个secret对象：
```go
// 检查该ServiceAccount对象引用的secrets对象存在，不存在则新建
func (e *TokensController) ensureReferencedToken(serviceAccount *v1.ServiceAccount) (bool, error) {
	// 首先确保serviceAccount.secrets字段值中的secret都存在
	if hasToken, err := e.hasReferencedToken(serviceAccount); err != nil {
		return false, err
	} else if hasToken {
		return false, nil
	}

	// 对api-server发起请求查找该serviceAccount对象
	serviceAccounts := e.client.CoreV1().ServiceAccounts(serviceAccount.Namespace)
	liveServiceAccount, err := serviceAccounts.Get(serviceAccount.Name, metav1.GetOptions{})
	// ...
	if liveServiceAccount.ResourceVersion != serviceAccount.ResourceVersion {
		return true, nil
	}

	// 如果是新建的ServiceAccount，则给ServiceAccount.secrets字段值添加个默认生成的secret对象
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Strategy.GenerateName(fmt.Sprintf("%s-token-", serviceAccount.Name)),
			Namespace: serviceAccount.Namespace,
			Annotations: map[string]string{
				v1.ServiceAccountNameKey: serviceAccount.Name, // 这里使用serviceAccount.Name来作为annotation
				v1.ServiceAccountUIDKey:  string(serviceAccount.UID), // 这里使用serviceAccount.UID来作为annotation
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
		Data: map[string][]byte{},
	}

	// 生成jwt token，该token是用私钥签名的
	token, err := e.token.GenerateToken(serviceaccount.LegacyClaims(*serviceAccount, *secret))
	// ...
	secret.Data[v1.ServiceAccountTokenKey] = []byte(token)
	secret.Data[v1.ServiceAccountNamespaceKey] = []byte(serviceAccount.Namespace)
	if e.rootCA != nil && len(e.rootCA) > 0 {
		secret.Data[v1.ServiceAccountRootCAKey] = e.rootCA
	}

	// 向api-server中创建该secret对象
	createdToken, err := e.client.CoreV1().Secrets(serviceAccount.Namespace).Create(secret)
	// ...
	// 写入LRU cache中
	e.updatedSecrets.Mutation(createdToken)

	err = clientretry.RetryOnConflict(clientretry.DefaultRetry, func() error {
		// ...
		// 把新建的secrets对象放入ServiceAccount.Secrets字段中，然后更新ServiceAccount对象
		liveServiceAccount.Secrets = append(liveServiceAccount.Secrets, v1.ObjectReference{Name: secret.Name})
		if _, err := serviceAccounts.Update(liveServiceAccount); err != nil {
			return err
		}
		// ...
	})

	// ...
}
```

所以，当ServiceAccount对象新建时，需要新建个新的Secret对象作为ServiceAccount对象的引用。业务代码还是比较简单的。


#### Secret的增删改查
当增删改查secret时，删除secret时同时需要删除serviceAccount对象下的secrets字段引用；

```go
func (e *TokensController) syncSecret() {
	// ...
	// 从LRU Cache中查找该secret
	secret, err := e.getSecret(secretInfo.namespace, secretInfo.name, secretInfo.uid, false)
	switch {
	case err != nil:
		klog.Error(err)
		retry = true
	case secret == nil:
		// 删除secret时：
		// 查找serviceAccount对象是否存在
		if sa, saErr := e.getServiceAccount(secretInfo.namespace, secretInfo.saName, secretInfo.saUID, false); saErr == nil && sa != nil {
			// 从service中删除其secret引用
			if err := clientretry.RetryOnConflict(RemoveTokenBackoff, func() error {
				return e.removeSecretReference(secretInfo.namespace, secretInfo.saName, secretInfo.saUID, secretInfo.name)
			}); err != nil {
				klog.Error(err)
			}
		}
	default:
		// 新建或更新secret时：
		// 查找serviceAccount对象是否存在
		sa, saErr := e.getServiceAccount(secretInfo.namespace, secretInfo.saName, secretInfo.saUID, true)
		switch {
		case saErr != nil:
			klog.Error(saErr)
			retry = true
		case sa == nil:
			// 如果serviceAccount都已经不存在，删除secret
			if retriable, err := e.deleteToken(secretInfo.namespace, secretInfo.name, secretInfo.uid); err != nil {
                // ...
			}
		default:
			// 新建或更新secret时，且serviceAccount存在时，查看是否需要更新secret中的ca/namespace/token字段值
			// 当然，新建secret时，肯定需要更新
			if retriable, err := e.generateTokenIfNeeded(sa, secret); err != nil {
				// ...
			}
		}
	}
}
```

所以，对kubernetes.io/service-account-token type的Secret增删改查的业务逻辑，也比较简单。重点是学习下官方golang代码编写和一些有关k8s api
的使用，对自己二次开发k8s大有裨益。


## 总结
本文主要学习TokensController是如何监听ServiceAccount对象和"kubernetes.io/service-account-token"类型Secret对象的增删改查，并做了相应的业务逻辑处理，
比如新建ServiceAccount时需要新建对应的Secret对象，删除ServiceAccount需要删除对应的Secret对象，以及新建Secret对象时，还需要给该Secret对象补上ca.crt/namespace/token
字段值，以及一些边界条件的处理逻辑等等。

同时，官方的TokensController代码编写规范，以及对k8s api的应用，边界条件的处理，以及使用了LRU Cache缓存等等，都值得在自己的项目里参(chao)考(xi)。


## 学习要点
**[tokens_controller.go L106](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/controller/serviceaccount/tokens_controller.go#L106) 使用了
LRU cache。


## 参考文献
**[为 Pod 配置服务账户](https://kubernetes.io/zh/docs/tasks/configure-pod-container/configure-service-account/)**
**[服务账号令牌 Secret](https://kubernetes.io/zh/docs/concepts/configuration/secret/#service-account-token-secrets)**


