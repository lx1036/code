package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" // import your used driver
	"github.com/gohouse/gorose/v2"
	"github.com/jinzhu/gorm"
	"reflect"
	"sync"
)

// Model Struct
type User struct {
	Id   int    `orm:"auto"`
	Name string `orm:"size(100)"`
}
type Post struct {
	Id    int    `orm:"auto"`
	Title string `orm:"size(100)"`
	User  *User  `orm:"rel(fk)"`
}

func beegoOrmDemo() {

	// set default database
	_ = orm.RegisterDataBase("default", "mysql", "root:root@tcp(127.0.0.1:3306)/beego_demo?charset=utf8mb4", 30)

	// register model
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Post))

	// create table
	_ = orm.RunSyncdb("default", false, true)

	o := orm.NewOrm()

	user := User{Name: "slene"}

	// insert
	id, err := o.Insert(&user)
	fmt.Printf("ID: %d, ERR: %v\n", id, err)

	// update
	user.Name = "astaxie"
	num, err := o.Update(&user)
	fmt.Printf("NUM: %d, ERR: %v\n", num, err)

	// read one
	u := User{Id: user.Id}
	err = o.Read(&u)
	fmt.Printf("ERR: %v\n", err)

	// delete
	//num, err = o.Delete(&u)
	//fmt.Printf("NUM: %d, ERR: %v\n", num, err)

	var posts []*Post
	qs := o.QueryTable("post")
	num, err = qs.Filter("User__Name", "astaxie").All(&posts)
	fmt.Printf("NUM: %d, ERR: %v\n", num, err)
}

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func gormDemo() {
	db, err := gorm.Open("mysql", "root:root@/orm?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic("connect db failed")
	}
	defer db.Close()

	//db.Set()
	//db.AutoMigrate(&Product{})
	//db.Create(&Product{Code: "L1212", Price: 1000})
	var product Product
	db.First(&product, 1)
	//db.First(&product, "code = ?", "L1212")
	//db.Model(&product).Update("Price", 2000)

	fmt.Println(product.Code)
}

func init() {
	engine, err = gorose.Open(&gorose.Config{
		Driver:          "mysql",
		Dsn:             "root:root@/orm?charset=utf8mb4&parseTime=True&loc=Local",
		SetMaxOpenConns: 10,
		SetMaxIdleConns: 1,
		Prefix:          "",
	})
	if err != nil {

	}
}

// `create database orm default character set utf8mb4 default collate utf8mb4_general_ci;`
func main() {
	//gormDemo()

	DBDemo()
}

var once sync.Once
var db gorose.IOrm
var engine *gorose.Engin
var err error

func DB() gorose.IOrm {
	//once.Do(func() {
	//	db = engine.NewOrm()
	//})
	db = engine.NewOrm()
	return db
}

func DBDemo() {
	results, err := DB().Query("select * from products where id = ?", 2)
	if err != nil {

	}
	fmt.Println(reflect.TypeOf(results).String())
	for _, result := range results {
		for column, value := range result {
			fmt.Println(column, value)
		}
	}

	/*affectedRows, err := DB().Execute("select * from products where id = ?", 2)
	if err != nil {

	}
	fmt.Println(affectedRows)*/

	result, err := db.Table("products").First()
	if err != nil {

	}
	for column, value := range result {
		fmt.Println(column, value)
	}

	products, err := db.Table("products").Get()
	for _, result := range products {
		for column, value := range result {
			fmt.Println(column, value)
		}
	}
}
