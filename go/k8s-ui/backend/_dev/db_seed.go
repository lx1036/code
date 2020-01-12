package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // import your used driver
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func main() {
	dbUrl := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&%s", "root", "root", "tcp(127.0.0.1:3306)", "k8s_ui", "Asia%2FShanghai")
	db, _ := sql.Open("mysql", dbUrl)
	defer db.Close()
	fmt.Println(dbUrl)
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback()
			panic(err)
		}
	}()

	file, _ := filepath.Abs("k8s-data.sql")
	fmt.Println(file)
	data, _ := ioutil.ReadFile(file)
	seeds := strings.Split(string(data), ";")
	fmt.Println(len(seeds))
	for _, query := range seeds {
		if len(query) != 0 {
			result, err := db.Exec(query)
			if err != nil {
				log.Println(query)
				log.Fatal(err)
			}
			rowsAffected, _ := result.RowsAffected()
			fmt.Println(rowsAffected)
		}
	}

	_ = tx.Commit()
}
