package orm_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
)

type User struct {
	Name, Email string
}

func TestConnection(t *testing.T) {
	connection := orm.NewConnection("sqlite3", GetDB(t))
	result, err := connection.Exec("INSERT INTO users(name,email) values('john doe','johndoe@acme.com'),('jane doe','jane.doe@acme.com');")
	if err != nil {
		t.Fatal(err)
	}
	if r, err := result.RowsAffected(); err != nil {
		t.Fatal(err)
	} else if r != 2 {
		t.Fatalf("2 records should have been created, got %d", r)
	} else {
		t.Log(result.LastInsertId())
	}

	// test query
	users := []*User{}
	err = connection.Select(&users, "SELECT users.name as Name, users.email as Email from users ORDER BY users.id ASC ;")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("%#v", users)
	}
	if expected, name := "john doe", users[0].Name; name != expected {
		t.Fatalf("users[0].Name should be '%s', got '%s'", expected, name)
	}
}

func GetDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: "./test/testdata/migrations/development.sqlite3",
	}
	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
