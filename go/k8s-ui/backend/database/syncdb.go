package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" // import your used driver
	_ "k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/util/logs"
)

func init() {
	err := beego.LoadAppConfig("ini", "conf/app.conf")
	if err != nil {
		panic(err)
	}
}

/**

 */
func main() {
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		panic(err)
	}

	dbUrl := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&%s",
		beego.AppConfig.String("DBUser"),
		beego.AppConfig.String("DBPassword"),
		beego.AppConfig.String("DBTns"),
		beego.AppConfig.String("DBName"),
		beego.AppConfig.String("DBLoc"),
	)

	logs.Info("db url: %s", dbUrl)

	err = orm.RegisterDataBase("default", "mysql", dbUrl)
	if err != nil {
		panic(err)
	}

	orm.RegisterModel()

	orm.RunCommand()
}
