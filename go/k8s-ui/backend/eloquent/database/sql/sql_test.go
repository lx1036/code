package sql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

// http://go-database-sql.org/index.html
func TestGoSqlDriverMysql(test *testing.T) {
	dsn := "root:root@tcp(127.0.0.1)/orm"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	//rows, err := db.Query(`select id, code from products where id = ?`, 1)
	statement, err := db.Prepare(`select id, code from products where id = ?`) // avoid sql injection
	if err != nil {
		panic(err)
	}

	rows, err := statement.Query(1)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		id   int
		code string
	)
	for rows.Next() {
		//columns, err := rows.Columns()
		//if err != nil {
		//	panic(err)
		//}

		err = rows.Scan(&id, &code)
		if err != nil {
			panic(err)
		}

		fmt.Println(id, code)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	stmt, err := db.Prepare("INSERT INTO products(code, price) VALUES(?, ?)")
	if err != nil {
		panic(err)
	}
	res, err := stmt.Exec("123", 100)
	if err != nil {
		panic(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	fmt.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}
