package lorm

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	DB *gorm.DB
)

func init() {
	var err error
	DB, err = gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/demo_k8s?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	//defer DB.Close() // open it, has errors, not know why
}
