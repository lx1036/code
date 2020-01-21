package eloquent

import "database/sql"

func Open(config DBConfig) *Connection {
	db, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {

	}

	return &Connection{DB: db}
}
