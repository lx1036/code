package lorm

import (
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
)

var (
	DB *gorm.DB
)

func init() {
	var err error
	DB, err = gorm.Open("mysql", "root:root@tcp(127.0.0.1)/k8s_ui?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
}
