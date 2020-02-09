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
		&models.App{},
		&models.Config{},
		&models.Group{},
		&models.Namespace{},
		&models.NamespaceUser{},
		&models.Notification{},
		&models.NotificationLog{},
		&models.User{},
	)

	db.Model(&models.APIKey{}).AddForeignKey("`group_id`", "`groups`(`id`)", "CASCADE", "CASCADE").
		AddForeignKey("`user_id`", "`users`(`id`)", "RESTRICT", "CASCADE") // use `group` quote identifier(preserved words)
	db.Model(&models.App{}).AddForeignKey("`user_id`", "`users`(`id`)", "CASCADE", "CASCADE").
		AddForeignKey("`namespace_id`", "`namespaces`(`id`)", "CASCADE", "CASCADE")
	db.Model(&models.NamespaceUser{}).AddForeignKey("`user_id`", "`users`(`id`)", "CASCADE", "CASCADE").
		AddForeignKey("`namespace_id`", "`namespaces`(`id`)", "CASCADE", "CASCADE")
	db.Model(&models.NotificationLog{}).AddForeignKey("`notification_id`", "`notifications`(`id`)", "CASCADE", "CASCADE").
		AddForeignKey("`user_id`", "`users`(`id`)", "CASCADE", "CASCADE")
	db.Model(&models.Notification{}).AddForeignKey("`from_user_id`", "`users`(`id`)", "CASCADE", "CASCADE")
}
