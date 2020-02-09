package cmd

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/romanyx/polluter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// go run main.go seed --data=database.yml

// Docs: https://medium.com/@romanyx90/testing-database-interactions-using-go-d9512b6bb449
// https://github.com/go-testfixtures/testfixtures
// https://github.com/khaiql/dbcleaner
// https://github.com/DATA-DOG/go-txdb
// https://github.com/romanyx/polluter
var (
	DB       *gorm.DB
	dataFile string
	content  string
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
			defer db.Close()

			p := polluter.New(polluter.MySQLEngine(db.DB()))

			if err := p.Pollute(strings.NewReader(content)); err != nil {
				panic(err)
			}

			/*tables := viper.AllKeys()
			fmt.Println(tables)

			for _, table := range tables {
				DB = DB.Table(table)

				values := viper.Get(table)
				fmt.Println(values, reflect.TypeOf(values))
				for _, value := range values.([]interface{}) {
					tmp, err := json.Marshal(value)
					if err != nil {
						panic(err)
					}
					fmt.Println(string(tmp))
				}
			}*/
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	seedCmd.PersistentFlags().StringVar(&dataFile, "data", "", "data yaml file path")
	_ = seedCmd.MarkFlagRequired("data")
}

func initConfig() {
	if dataFile != "" {
		filename, _ := filepath.Abs("./database.yml")
		data, _ := ioutil.ReadFile(filename)

		content = string(data)
		// Use config file from the flag.
		viper.SetConfigFile(dataFile)
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
