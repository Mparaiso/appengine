package orm

import (
	"database/sql"
	"fmt"

	"reflect"

	"github.com/mparaiso/go-orm/tools"
)

type ConnectionOptions struct {
	Logger
	IgnoreMissingFields bool
}

// Connection is a database connection.
// Please use NewConnectionto create a Connection.
type Connection struct {
	db         *sql.DB
	driverName string
	Options    *ConnectionOptions
}

// NewConnection creates an new Connection
func NewConnection(driverName string, DB *sql.DB) *Connection {
	return NewConnectionWithOptions(driverName, DB, &ConnectionOptions{})
}

// NewConnectionWithOptions creates an new connection with optional settings such as Logging.
func NewConnectionWithOptions(driverName string, DB *sql.DB, options *ConnectionOptions) *Connection {
	return &Connection{DB, driverName, options}
}

// DriverName returns the DriverName
func (connection *Connection) DriverName() string {
	return connection.driverName
}

// DB returns Go standard *sql.DB type
func (connection *Connection) DB() *sql.DB {
	return connection.db
}

// Exec will execute a query like INSERT,UPDATE,DELETE.
func (connection *Connection) Exec(query string, parameters ...interface{}) (sql.Result, error) {
	defer connection.log(append([]interface{}{query}, parameters...)...)
	return connection.DB().Exec(query, parameters...)
}

// Select with fetch multiple records.
func (connection *Connection) Select(records interface{}, query string, parameters ...interface{}) error {
	defer connection.log(append([]interface{}{query}, parameters...)...)

	rows, err := connection.db.Query(query, parameters...)
	if err != nil {
		return err
	}
	err = tools.MapRowsToSliceOfStruct(rows, records, true)

	return err
}

// Get will fetch a single record.
// expects record to be a pointer to a struct
// Exemple:
//    user := new(User)
//    err := connection.get(user,"SELECT * from users WHERE users.id = ?",1)
//
func (connection *Connection) Get(record interface{}, query string, parameters ...interface{}) error {
	// make a slice from the record type
	// pass a pointer to that slice to connection.Select
	// if the slice's length == 1 , put back the first value of that
	// slice in the record value.
	if reflect.TypeOf(record).Kind() != reflect.Ptr {
		return NotAPointerError(fmt.Sprintf("Expecting a pointer, got %#v", record))
	}
	recordValue := reflect.ValueOf(record)
	recordType := recordValue.Type()
	sliceOfRecords := reflect.MakeSlice(reflect.SliceOf(recordType), 0, 1)
	pointerOfSliceOfRecords := reflect.New(sliceOfRecords.Type())
	pointerOfSliceOfRecords.Elem().Set(sliceOfRecords)
	//
	err := connection.Select(pointerOfSliceOfRecords.Interface(), query, parameters...)
	if err != nil {
		return err
	}
	if pointerOfSliceOfRecords.Elem().Len() >= 1 {
		recordValue.Elem().Set(reflect.Indirect(pointerOfSliceOfRecords).Index(0).Elem())
	} else {
		return sql.ErrNoRows
	}
	return nil
}

func (connection *Connection) log(messages ...interface{}) {
	if connection.Options.Logger != nil {
		connection.Options.Logger.Log(messages...)
	}
}

func (connection *Connection) BeginTransaction() (*Transaction, error) {
	defer connection.log("Begin transaction")
	transaction, err := connection.DB().Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{Logger: connection.Options.Logger, Tx: transaction}, nil
}

func (connection *Connection) Close() error {
	return connection.db.Close()
}
