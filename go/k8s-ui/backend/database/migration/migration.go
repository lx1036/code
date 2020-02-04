package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"k8s-lx1036/k8s-ui/backend/models"
)

func main() {
	dbName := "demo_k8s"
	db, err := gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		switch err.(type) {
		case *mysql.MySQLError:
			_, _ = db.DB().Exec(fmt.Sprintf(`create database %s;`, dbName))
		default:
			panic(err)
		}
	}
	defer db.Close()

	db.Exec(`drop database if exists demo_k8s;`).Exec("create database demo_k8s;").Exec("use demo_k8s;")
	/*if err != nil {
		panic(err)
	}
	_, err = db.DB().Exec(fmt.Sprintf(`create database %s;`, dbName))
	if err != nil {
		panic(err)
	}
	_, err = db.DB().Exec(fmt.Sprintf(`use %s;`, dbName))
	if err != nil {
		panic(err)
	}

	db.Exec("")*/

	db, _ = gorm.Open("mysql", fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbName))
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		&models.APIKey{},
		&models.Group{},
		&models.User{},
	)

	db.Debug()
	db.Model(&models.APIKey{}).AddForeignKey("`group_id`", "`groups`(`id`)", "RESTRICT", "RESTRICT"). // use `group` quote identifier(preserved words)
														AddForeignKey("`user_id`", "`users`(`id`)", "RESTRICT", "RESTRICT")
}
