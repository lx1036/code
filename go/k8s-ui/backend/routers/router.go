package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/controllers/config"
	"k8s-lx1036/k8s-ui/backend/controllers/deployment"
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
				&deployment.DeploymentController{},
			),
		),
	)

	nsWithoutApp := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/configs/base",
			beego.NSInclude(
				&config.BaseConfigController{},
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
		beego.NSNamespace("/namespaces/:namespaceid([0-9]+)/users",
			beego.NSInclude(
				&permission.NamespaceUserController{},
			),
		),
	)

	beego.AddNamespace(withApp)
	beego.AddNamespace(nsWithoutApp)
	beego.AddNamespace(nsWithNamespace)
	beego.AddNamespace(nsWithOpenAPI)
}
