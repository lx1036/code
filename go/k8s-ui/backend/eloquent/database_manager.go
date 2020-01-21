package eloquent

import "database/sql"

type DatabaseManager struct {
	db                *sql.DB
	connections       []Connection
	connectionFactory ConnectionFactory
}

type DBConfig struct {
	Driver          string
	Dsn             string
	SetMaxOpenConns int
	SetMaxIdleConns int
	Prefix          string
}

func (manager *DatabaseManager) Open(config DBConfig) {
	db, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {
		return
	}

	err = db.Ping()
	if err != nil {
		return
	}

	db.SetMaxOpenConns(config.SetMaxOpenConns)
	db.SetMaxIdleConns(config.SetMaxIdleConns)

	manager.db = db
}

func (manager *DatabaseManager) Connection(name string) Connection {

	return manager.connectionFactory.Make(config, name)
}

func New() *Manager {
	return &Manager{}
}
