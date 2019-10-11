package routes

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"k8s-lx1036/wayne/backend/controllers/cronjob"
	"k8s-lx1036/wayne/backend/controllers/kubernetes"
	"k8s-lx1036/wayne/backend/controllers/permission"
	"path"
)


func init()  {
	if beego.BConfig.RunMode == "dev" && path.Base(beego.AppPath) == "_build" {
		beego.AppPath = path.Join(path.Dir(beego.AppPath), "src/backend")
	}

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))


	beego.AddNamespace(beego.NewNamespace("/api/v1",
		// 路由中携带appid
		beego.NSNamespace("/apps/:appid([0-9]+)/users",
			beego.NSInclude(
				&permission.AppUserController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/configmaps",
			beego.NSInclude(
				&configmap.ConfigMapController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/configmaps/tpls",
			beego.NSInclude(
				&configmap.ConfigMapTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/cronjobs",
			beego.NSInclude(
				&cronjob.CronjobController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/cronjobs/tpls",
			beego.NSInclude(
				&cronjob.CronjobTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/deployments",
			beego.NSInclude(
				&deployment.DeploymentController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/deployments/tpls",
			beego.NSInclude(
				&deployment.DeploymentTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/statefulsets",
			beego.NSInclude(
				&statefulset.StatefulsetController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/statefulsets/tpls",
			beego.NSInclude(
				&statefulset.StatefulsetTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/daemonsets",
			beego.NSInclude(
				&daemonset.DaemonSetController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/daemonsets/tpls",
			beego.NSInclude(
				&daemonset.DaemonSetTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/persistentvolumeclaims",
			beego.NSInclude(
				&pvc.PersistentVolumeClaimController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/persistentvolumeclaims/tpls",
			beego.NSInclude(
				&pvc.PersistentVolumeClaimTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/secrets",
			beego.NSInclude(
				&secret.SecretController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/secrets/tpls",
			beego.NSInclude(
				&secret.SecretTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/webhooks",
			beego.NSInclude(
				&webhook.WebHookController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/apikeys",
			beego.NSInclude(
				&apikey.ApiKeyController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/ingresses",
			beego.NSInclude(
				&ingress.IngressController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/ingresses/tpls",
			beego.NSInclude(
				&ingress.IngressTplController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/hpas",
			beego.NSInclude(
				&hpa.HPAController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/hpas/tpls",
			beego.NSInclude(
				&hpa.HPATplController{},
			),
		),
	))

	// kubernetes resource
	beego.AddNamespace(beego.NewNamespace("/api/v1",
		beego.NSRouter("/kubernetes/pods/statistics", &kpod.KubePodController{}, "get:PodStatistics"),

		beego.NSNamespace("/kubernetes/persistentvolumes",
			beego.NSInclude(
				&kpv.KubePersistentVolumeController{},
			),
		),
		beego.NSNamespace("/kubernetes/persistentvolumes/robin",
			beego.NSInclude(
				&kpv.RobinPersistentVolumeController{},
			),
		),
		beego.NSNamespace("/kubernetes/namespaces",
			beego.NSInclude(
				&kubernetes.KubernetesNamespaceController{},
			),
		),
		beego.NSNamespace("/kubernetes/nodes",
			beego.NSInclude(
				&knode.KubeNodeController{},
			),
		),
	))

	beego.AddNamespace(beego.NewNamespace("/api/v1",
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/cronjobs",
			beego.NSInclude(
				&kcronjob.KubeCronjobController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/deployments",
			beego.NSInclude(
				&kdeployment.KubeDeploymentController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/statefulsets",
			beego.NSInclude(
				&kstatefulset.KubeStatefulsetController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/daemonsets",
			beego.NSInclude(
				&kdaemonset.KubeDaemonSetController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/configmaps",
			beego.NSInclude(
				&kconfigmap.KubeConfigMapController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/services",
			beego.NSInclude(
				&kservice.KubeServiceController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/ingresses",
			beego.NSInclude(
				&kingress.KubeIngressController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/hpas",
			beego.NSInclude(
				&khpa.KubeHPAController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/secrets",
			beego.NSInclude(
				&ksecret.KubeSecretController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/persistentvolumeclaims",
			beego.NSInclude(
				&kpvc.KubePersistentVolumeClaimController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/persistentvolumeclaims/robin",
			beego.NSInclude(
				&kpvc.RobinPersistentVolumeClaimController{},
			),
		),

		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/jobs",
			beego.NSInclude(
				&kjob.KubeJobController{},
			),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/pods",
			beego.NSInclude(
				&kpod.KubePodController{}),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/events",
			beego.NSInclude(
				&kevent.KubeEventController{}),
		),
		beego.NSNamespace("/kubernetes/apps/:appid([0-9]+)/podlogs",
			beego.NSInclude(
				&klog.KubeLogController{}),
		),
	))

	beego.AddNamespace(beego.NewNamespace("/api/v1",
		// 路由中携带namespaceid
		beego.NSNamespace("/namespaces/:namespaceid([0-9]+)/apps",
			beego.NSInclude(
				&app.AppController{},
			),
		),
		beego.NSNamespace("/namespaces/:namespaceid([0-9]+)/webhooks",
			beego.NSInclude(
				&webhook.WebHookController{},
			),
		),
		beego.NSNamespace("/namespaces/:namespaceid([0-9]+)/apikeys",
			beego.NSInclude(
				&apikey.ApiKeyController{},
			),
		),
		beego.NSNamespace("/namespaces/:namespaceid([0-9]+)/users",
			beego.NSInclude(
				&permission.NamespaceUserController{},
			),
		),
		beego.NSNamespace("/namespaces/:namespaceid([0-9]+)/bills",
			beego.NSInclude(
				&bill.BillController{},
			),
		),
	))


	beego.AddNamespace(beego.NewNamespace("/api/v1",
		// 路由中不携带任何id
		beego.NSNamespace("/configs",
			beego.NSInclude(
				&config.ConfigController{},
			),
		),
		beego.NSNamespace("/configs/base",
			beego.NSInclude(
				&config.BaseConfigController{},
			),
		),
		beego.NSRouter("/apps/statistics", &app.AppController{}, "get:AppStatistics"),
		beego.NSNamespace("/clusters",
			beego.NSInclude(
				&cluster.ClusterController{},
			),
		),
		beego.NSNamespace("/auditlogs",
			beego.NSInclude(
				&auditlog.AuditLogController{},
			),
		),
		beego.NSNamespace("/notifications",
			beego.NSInclude(
				&notification.NotificationController{},
			),
		),
		beego.NSNamespace("/namespaces",
			beego.NSInclude(
				&namespace.NamespaceController{},
			),
		),
		beego.NSNamespace("/apps/stars",
			beego.NSInclude(
				&appstarred.AppStarredController{},
			),
		),
		beego.NSNamespace("/publish",
			beego.NSInclude(
				&publish.PublishController{},
			),
		),
		beego.NSNamespace("/publishstatus",
			beego.NSInclude(
				&publishstatus.PublishStatusController{},
			),
		),
		beego.NSNamespace("/users",
			beego.NSInclude(
				&permission.UserController{}),
		),
		beego.NSNamespace("/groups",
			beego.NSInclude(
				&permission.GroupController{},
			),
		),
		beego.NSNamespace("/permissions",
			beego.NSInclude(
				&permission.PermissionController{},
			),
		),
	))

	beego.AddNamespace(beego.NewNamespace("/api/v1",
		beego.NSNamespace("/apps/:appid([0-9]+)/_proxy/clusters/:cluster/namespaces/:namespace/:kind",
			beego.NSInclude(
				&proxy.KubeProxyController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/_proxy/clusters/:cluster/customresourcedefinitions",
			beego.NSInclude(
				&kcrd.KubeCRDController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/_proxy/clusters/:cluster/apis/:group/:version/namespaces/:namespace/:kind",
			beego.NSInclude(
				&kcrd.KubeCustomCRDController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/_proxy/clusters/:cluster/apis/:group/:version/:kind",
			beego.NSInclude(
				&kcrd.KubeCustomCRDController{},
			),
		),
		beego.NSNamespace("/apps/:appid([0-9]+)/_proxy/clusters/:cluster/:kind",
			beego.NSInclude(
				&proxy.KubeProxyController{},
			),
		),
	))

}









