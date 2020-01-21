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

func (manager *DatabaseManager) Open(config DBConfig) *DatabaseManager {
	db, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {
		return nil
	}

	err = db.Ping()
	if err != nil {
		return nil
	}

	db.SetMaxOpenConns(config.SetMaxOpenConns)
	db.SetMaxIdleConns(config.SetMaxIdleConns)

	manager.db = db

	return manager
}

func (manager *DatabaseManager) Connection( /*name string*/ ) *Connection {
	connection := Connection{DB: manager.db}
	return &connection
	//return &Connection{
	//	DB: manager.db,
	//}
	//return manager.connectionFactory.Make(config, name)
}

func New() *DatabaseManager {
	return &DatabaseManager{}
}
