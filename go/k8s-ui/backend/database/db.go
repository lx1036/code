package database

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

var (
	DB *gorm.DB
)

func InitDb() *gorm.DB {
	var err error
	DB, err = gorm.Open("mysql", GetDBUrl())
	if err != nil {
		panic(err)
	}
	return DB
}

func GetDBUrl() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		viper.GetString("default.DbUser"),
		viper.GetString("default.DbPassword"),
		viper.GetString("default.DbHost"),
		viper.GetInt("default.DbPort"),
		viper.GetString("default.DbName"),
		viper.GetString("default.DbLoc"),
	)
}
