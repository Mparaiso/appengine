package orm

import (
	"database/sql"
)

type Transaction struct {
	*sql.Tx
	Logger Logger
}

func (transaction *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	defer transaction.log(append([]interface{}{query}, args...)...)
	return transaction.Tx.Exec(query, args...)
}

func (transaction *Transaction) RollBack() error {
	defer transaction.log("Rollback Transaction.")
	return transaction.Tx.Rollback()
}

func (transaction *Transaction) Commit() error {
	defer transaction.log("Commit Transaction.")
	return transaction.Tx.Commit()
}

func (transaction *Transaction) log(messages ...interface{}) {
	if transaction.Logger != nil {
		transaction.Logger.Log(messages...)
	}
}
