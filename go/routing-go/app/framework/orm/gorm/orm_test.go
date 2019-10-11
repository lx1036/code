package gorm

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "log"
    "testing"
)

type Account struct {
    gorm.Model
    Name string
    Type string
    Source string `json:"source"`
    //CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func TestGorm(test *testing.T) {
    db, err := gorm.Open("mysql", "testing:testing@(localhost:3306)/rightcapital__lx1036?parseTime=true")
    if err != nil {
        panic("failed to connect db")
    }
    defer db.Close()

    var account Account
    db.Find(&account, "person_id=?", "3")
    log.Println(account)
}
