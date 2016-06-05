package datamapper

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type ConnectionOptions struct {
	Logger
}

// Connection is a database connection.
// Please use NewConnectionto create a Connection.
type Connection struct {
	db         *sqlx.DB
	driverName string
	Options    *ConnectionOptions
}

// NewConnection creates an new Connection
func NewConnection(driverName string, DB *sql.DB) *Connection {
	return NewConnectionWithOptions(driverName, DB, &ConnectionOptions{})
}

func NewConnectionWithOptions(driverName string, DB *sql.DB, options *ConnectionOptions) *Connection {
	return &Connection{sqlx.NewDb(DB, driverName), driverName, options}
}

// DriverName returns the DriverName
func (connection *Connection) DriverName() string {
	return connection.driverName
}

// DB returns Go standard *sql.DB type
func (connection *Connection) DB() *sql.DB {
	return connection.db.DB
}

// Exec will execute a query like INSERT,UPDATE,DELETE.
func (connection *Connection) Exec(query string, parameters ...interface{}) (sql.Result, error) {
	defer connection.log(append([]interface{}{query}, parameters...))
	return connection.db.Unsafe().Exec(query, parameters...)
}

// Select with fetch multiple records.
func (connection *Connection) Select(records interface{}, query string, parameters ...interface{}) error {
	defer connection.log(append([]interface{}{query}, parameters...))
	return connection.db.Unsafe().Select(records, query, parameters...)

}

// Get will fetch a single record.
func (connection *Connection) Get(record interface{}, query string, parameters ...interface{}) error {
	defer connection.log(append([]interface{}{query}, parameters...))
	return connection.db.Unsafe().Get(record, query, parameters...)
}

func (connection *Connection) log(messages ...interface{}) {
	if connection.Options.Logger != nil {
		connection.Options.Logger.Log(messages...)
	}
}

func (connection *Connection) BeginTransaction() (*sql.Tx, error) {
	defer connection.log("Begin transaction")
	return connection.db.Begin()
}



func (connection *Connection) Close() error {
	return connection.db.Close()
}
