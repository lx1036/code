package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/romanyx/polluter"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

var input = `
users:
  - id: 1
    name: "admin"
    password: "password"
    salt: "abc123"
    email: "admin@example.com"
    display: "Admin"
    comment: "This is a super user for k8s ui"
    type: 1
    admin: 1
    last_login: "2020-02-03 00:00:00"
    last_ip: "127.0.0.1"
  - id: 2
    name: "developer"
    password: "password"
    salt: "abc123"
    email: "developer@example.com"
    display: "Developer"
    comment: "This is a developer user for k8s ui"
    type: 0
    admin: 0
    last_login: "2020-02-03 00:00:00"
    last_ip: "127.0.0.1"
`

func TestSeed(test *testing.T) {
	dbName := "demo_k8s"
	db, err := gorm.Open("mysql", fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	filename, _ := filepath.Abs("./database.yml")
	content, _ := ioutil.ReadFile(filename)

	p := polluter.New(polluter.MySQLEngine(db.DB()))

	if err := p.Pollute(strings.NewReader(string(content))); err != nil {
		panic(err)
	}
}
