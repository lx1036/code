package lorm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
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

func GetDBUrl() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("DB_USERNAME"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_HOST"),
		viper.GetInt("DB_PORT"),
		viper.GetString("DB_NAME"),
	)
}
