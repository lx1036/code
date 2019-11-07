package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" // import your used driver
)

func init() {
	err := beego.LoadAppConfig("ini", "backend/conf/app.conf")
	if err != nil {
		panic(err)
	}
}

func main() {
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		panic(err)
	}

	dbUrl := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&",
		beego.AppConfig.String("DBUser"),
		beego.AppConfig.String("DBPasswd"),
		beego.AppConfig.String("DBTns"),
		beego.AppConfig.String("DBName"),
	)

	err = orm.RegisterDataBase("default", "mysql", dbUrl)
	if err != nil {
		panic(err)
	}

	orm.RunCommand()
}
