package gorm

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"testing"
)

type Account struct {
	gorm.Model
	Name   string
	Type   string
	Source string `json:"source"`
	//CreatedAt time.Time `db:"created_at" json:"created_at"`
}

const (
	TableNameApiKey = "api_keys"
)

type APIKey struct {
	Id    uint   `gorm:"auto_increment;column:id;type:bigint;size:20;primary_key;not null"`
	Name  string `gorm:"column:name;size:128;not null;index:api_key_name;default:''"`
	Token string `gorm:"column:token;type:longtext;not null"`
}

func (APIKey) TableName() string {
	return TableNameApiKey
}

type User struct {
	gorm.Model
	Emails []Email `gorm:"ForeignKey:UserID;AssociationForeignKey:ID"`
}

type Email struct {
	gorm.Model
	Email  string
	UserID uint
}

func TestGorm(test *testing.T) {

	dbName := "demo1"
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

	db.Exec(`drop database if exists demo1;`).Exec("create database demo1;").Exec("use demo1;")

	db, _ = gorm.Open("mysql", fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbName))

	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		&User{},
		&Email{},
	)

	var user User
	var emails []Email
	db.Model(&user).Related(&emails)

	//var account Account
	//db.Find(&account, "person_id=?", "3")
	//log.Println(account)
}
