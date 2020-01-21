package eloquent

import (
	"database/sql"
	"fmt"
	"k8s-lx1036/k8s-ui/backend/eloquent/grammar"
	"time"
)

type Connection struct {
	DB *sql.DB
}

type Data map[string]interface{}

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
func (connection *Connection) Select(query string, bindings []interface{}) []Data {
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

		result = append(result, data)
	}

	//connection.logQuery(query, bindings,)
	fmt.Println(time.Since(start).Milliseconds())

	return result
}
