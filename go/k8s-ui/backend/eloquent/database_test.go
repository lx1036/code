package eloquent

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestManager(test *testing.T) {
	/*connection := New().Open(DBConfig{
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
	}*/
}

var connection *Connection

func init() {
	connection = initDb()
}

func initDb() *Connection {
	database := "orm"
	dsn := "root:root@tcp(127.0.0.1:3306)/"
	driver := "mysql"
	connection, err := Open(DBConfig{
		Driver: driver,
		Dsn:    dsn + "?charset=utf8&parseTime=true",
	})
	if err != nil {
		panic(err)
	}
	err = connection.Statement(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s;`, database))

	if err != nil {
		panic(err)
	}

	connection, err = Open(DBConfig{
		Driver: driver,
		Dsn:    dsn + fmt.Sprintf("%s?charset=utf8&parseTime=true", database),
	})
	if err != nil {
		panic(err)
	}

	err = connection.Statement(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		panic(err)
	}
	err = connection.Statement(`
CREATE TABLE users (
  id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name varchar(255) DEFAULT NULL,
  gender varchar(255) DEFAULT NULL,
  addr varchar(255) DEFAULT NULL,
  balance decimal(15,4) DEFAULT '0.0000',
  birth_date date DEFAULT NULL,
  created_at timestamp NULL DEFAULT NULL,
  updated_at timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
) ENGINE=InnoDB;
`)
	if err != nil {
		panic(err)
	}

	return connection
}

type User struct {
	Id        int64  `torm:"primary_key;column:id"`
	Name      string `torm:"column:name"`
	Gender    string `torm:"column:gender"`
	Addr      string
	BirthDate string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user *User) TableName() string {
	return "users"
}

func TestConnectionSelectOne(t *testing.T) {
	var user User
	_, _, err := connection.Insert("insert into users (name, gender) values (?, ?)", "Andrew", "M")
	if err != nil {
		t.Error(err)
	}

	err = connection.SelectOne("select * from users where gender = ?", []interface{}{"M"}, &user)
	if err != nil {
		t.Error(err)
	}

	if user.Gender != "M" {
		t.Error("Expect: user's gender should be ", "M")
	}

	if user.Name != "Andrew" {
		t.Error("Expect: user's name should be ", "Andrew")
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
		Dsn:             "root:root@tcp(127.0.0.1:3306)/orm?charset=utf8mb4&parseTime=True&loc=Local",
		SetMaxOpenConns: 10,
		SetMaxIdleConns: 1,
		Prefix:          "",
	}
	connection, _ := Open(config)

	var products []Product
	err := connection.Table("products").Where("id", 1).Get(&products)
	if err != nil {
		panic(err)
	}

	for _, product := range products {
		fmt.Println(product.Code)
	}
}
