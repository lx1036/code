package eloquent

import (
	"database/sql"
	"fmt"
	"k8s-lx1036/k8s-ui/backend/eloquent/grammar"
	"reflect"
	"time"
)

type Connection struct {
	DB *sql.DB
}

type Data map[string]interface{}

// Statement Execute an SQL statement and return the boolean result.
func (connection *Connection) Statement(query string, bindings ...interface{}) error {
	statement, err := connection.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(bindings)

	return err
}

func (connection *Connection) Table(table string) *Builder {
	return connection.Query().From(table)
}

func (connection *Connection) Query() *Builder {
	return NewBuilder(connection, connection.GetGrammar())
}

func (connection *Connection) GetGrammar() grammar.Grammar {
	return &grammar.MysqlGrammar{}
}

// Run a SQL statement and log its execution context.
func (connection *Connection) Select(query string, bindings []interface{}, dest interface{}) error {
	start := time.Now()

	statement, err := connection.DB.Prepare(query)
	if err != nil {

	}
	defer statement.Close()
	rows, err := statement.Query(bindings...)
	if err != nil {

	}
	defer rows.Close()

	//result := connection.runQuery()
	columns, err := rows.Columns()
	if err != nil {

	}

	count := len(columns)

	var result []Data
	//var resultValue reflect.Value

	results := reflect.Indirect(reflect.ValueOf(dest))
	resultValue := results

	for rows.Next() {
		values := make([]interface{}, count)
		args := make([]interface{}, count)
		for i := 0; i < count; i++ {
			args[i] = &values[i]
		}

		err := rows.Scan(args...)
		if err != nil {

		}

		var data = Data{}
		for key, column := range columns {
			value := values[key]
			data[column] = value
		}

		results.Set(reflect.Append(results, resultValue))

		result = append(result, data)
	}

	//connection.logQuery(query, bindings,)
	fmt.Println(time.Since(start).Milliseconds())

	return nil
}

func (connection *Connection) SelectOne(query string, bindings []interface{}, dest interface{}) error {
	return connection.Select(query, bindings, dest)
}

func (connection *Connection) Insert(query string, args ...interface{}) (int64, int64, error) {
	return connection.affectingStatement(query, args...)
}

func (connection *Connection) affectingStatement(query string, args ...interface{}) (int64, int64, error) {
	stmt, err := connection.DB.Prepare(query)

	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)

	if err != nil {
		return 0, 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, affected, err
	}

	insertId, err := result.LastInsertId()
	return insertId, affected, err
}
