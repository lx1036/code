package eloquent

import (
	"fmt"
	"testing"
	"time"
)

func TestManager(test *testing.T) {
	connection := New().Open(DBConfig{
		Driver:          "mysql",
		Dsn:             "root:root@/orm?charset=utf8mb4&parseTime=True&loc=Local",
		SetMaxOpenConns: 10,
		SetMaxIdleConns: 1,
		Prefix:          "",
	}).Connection()

	collection := connection.Select("select * from ? where id = ?", []interface{}{"products", "1"})
	for _, row := range collection {
		for column, value := range row {
			fmt.Println(column, value)
		}
	}
}

type Product struct {
	Id        string `lorm:"column:id"`
	Code      string `lorm:"column:code"`
	Price     int    `lorm:"column:price"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// https://github.com/thinkoner/torm/blob/master/README.md
func TestBasic(test *testing.T) {
	config := DBConfig{
		Driver:          "mysql",
		Dsn:             "root:root@/orm?charset=utf8mb4&parseTime=True&loc=Local",
		SetMaxOpenConns: 10,
		SetMaxIdleConns: 1,
		Prefix:          "",
	}
	connection := Open(config)

	var products []Product
	connection.Table("products").Where("id", 1).Get(&products)
}
