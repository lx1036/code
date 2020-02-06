package cmd

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	DB       *gorm.DB
	dataFile string
	seedCmd  = &cobra.Command{
		Use:   "seed",
		Short: "seed data to db",
		Run: func(cmd *cobra.Command, args []string) {
			dbName := "demo_k8s"
			db, err := gorm.Open("mysql", fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbName))
			if err != nil {
				switch err.(type) {
				case *mysql.MySQLError:
					_, _ = db.DB().Exec(fmt.Sprintf(`create database %s;`, dbName))
				default:
					panic(err)
				}
			}

			DB = db

			tables := viper.AllKeys()
			fmt.Println(tables)

			for _, table := range tables {
				DB = DB.Table(table)

				for _, value := range viper.Get(table).([]interface{}) {
					fmt.Println(value)
				}
			}
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	seedCmd.PersistentFlags().StringVar(&dataFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	_ = seedCmd.MarkFlagRequired("config")

}

func initConfig() {
	if dataFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(dataFile)
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
