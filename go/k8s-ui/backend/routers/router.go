package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"k8s-lx1036/k8s-ui/backend/controllers"
	"k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/controllers/kubernetes"
	"k8s-lx1036/k8s-ui/backend/controllers/openapi"
	"k8s-lx1036/k8s-ui/backend/controllers/permission"
	"path"
)

func init() {
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
	/**
	Auth
	*/
	beego.Include(&auth.AuthController{})

	withApp := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/apps/:appid([0-9]+)/users",
			beego.NSInclude(
				&permission.AppUserController{},
			),
		),

		beego.NSNamespace("/apps/:appid([0-9]+)/deployments",
			beego.NSInclude(
				&controllers.DeploymentController{},
			),
		),
	)

	nsWithoutApp := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/configs/base",
			beego.NSInclude(
				&controllers.BaseConfigController{},
			),
		),
		beego.NSNamespace("/notifications",
			beego.NSInclude(
				&controllers.NotificationController{},
			),
		),
		beego.NSNamespace("/users",
			beego.NSInclude(
				&permission.UserController{},
			),
		),
		beego.NSRouter("/apps/statistics", &controllers.AppController{}, "get:AppStatistics"),
		beego.NSNamespace("/namespaces",
			beego.NSInclude(
				&controllers.NamespaceController{},
			),
		),
	)

	nsWithOpenAPI := beego.NewNamespace("/openapi/v1",
		beego.NSNamespace("/gateway/action",
			beego.NSInclude(
				&openapi.OpenAPIController{}),
		),
	)

	nsWithNamespace := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/namespaces/:namespaceId([0-9]+)/apps",
			beego.NSInclude(
				&controllers.AppController{},
			),
		),
		beego.NSNamespace("/namespaces/:namespaceId([0-9]+)/users",
			beego.NSInclude(
				&permission.NamespaceUserController{},
			),
		),
	)

	nsWithKubernetes := beego.NewNamespace("/api/v1",
		beego.NSRouter("/kubernetes/pods/statistics", &kubernetes.KubePodController{}, "get:PodStatistics"),
		beego.NSNamespace("/kubernetes/namespaces",
			beego.NSInclude(
				&kubernetes.KubeNamespaceController{},
			),
		),
	)

	beego.AddNamespace(withApp)
	beego.AddNamespace(nsWithoutApp)
	beego.AddNamespace(nsWithNamespace)
	beego.AddNamespace(nsWithOpenAPI)
	beego.AddNamespace(nsWithKubernetes)
}
