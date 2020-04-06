package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var (
	number int
)

type Account struct {
	Id     int    `json:"id"`
	Number string `json:"number"`
	Name   string `json:"name"`
}

func (Account) TableName() string {
	return "account" // not accounts to check this function
}

func main() {
	viper.AutomaticEnv()

	var rootCmd = &cobra.Command{
		Use:    "ProjectName",
		Run:    run,
		PreRun: preRun,
	}

	_ = rootCmd.Execute()
}

func preRun(cmd *cobra.Command, args []string) {
	number = viper.GetInt("NUMBER")
}

func run(cmd *cobra.Command, args []string) {
	router := gin.Default()
	router.GET("/hello", Hello())
	router.GET("/test", Test())
	var port int
	if port = viper.GetInt("PORT"); port == 0 {
		port = 8080
	}
	router.Run(fmt.Sprintf(":%d", port))
}
func Hello() gin.HandlerFunc {
	return func(context *gin.Context) {
		db, err := gorm.Open("mysql", GetDBUrl())
		if err != nil {
			panic(err)
		}
		defer db.Close()

		var account Account
		err = db.Where("number=?", number).First(&account).Error
		if err != nil {

		}
		context.JSON(http.StatusOK, gin.H{
			"data": account,
		})
	}
}
func Test() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"data": "test",
		})
	}
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
