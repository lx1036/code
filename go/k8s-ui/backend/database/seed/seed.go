package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func main() {
	dbName := "demo_k8s"
	db, err := gorm.Open("mysql", fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbName))
	//db, err := gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		switch err.(type) {
		case *mysql.MySQLError:
			_, _ = db.DB().Exec(fmt.Sprintf(`create database %s;`, dbName))
		default:
			panic(err)
		}
	}
	defer db.Close()

}
