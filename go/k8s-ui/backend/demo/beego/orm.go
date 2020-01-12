package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" // import your used driver
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

func init() {
	// set default database
	_ = orm.RegisterDataBase("default", "mysql", "root:root@tcp(127.0.0.1:3306)/beego_demo?charset=utf8mb4", 30)

	// register model
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Post))

	// create table
	_ = orm.RunSyncdb("default", false, true)
}

func main() {
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
