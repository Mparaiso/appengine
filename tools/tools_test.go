package tools_test

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/go-orm/tools"
	"github.com/rubenv/sql-migrate"
)

import (
	"testing"
)

type User struct {
	Name             string
	Email            string
	Created, Updated time.Time
	*UserInfo
}
type UserInfo struct {
	Url string
}

func TestMapRowsToSliceOfStruct(t *testing.T) {
	db := GetConnection(t)
	users := []*User{
		{Name: "John Doe", Email: "john.doe@acme.com"},
		{Name: "Jane Doe", Email: "jane.doe@acme.com"},
	}
	for _, user := range users {
		_, err := db.Exec("INSERT INTO users(name,email) values(?,?)", user.Name, user.Email)
		if err != nil {
			t.Fatal(err)
		}
	}
	result := []*User{}
	rows, err := db.Query("SELECT name as Name,email as Email,created as Created from users ORDER BY users.id ;")
	if err != nil {
		t.Fatal(err)
	}
	err = tools.MapRowsToSliceOfStruct(rows, &result, true)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 2 {
		t.Fatalf("Result length should be 2, got %d", l)
	}
}

func TestMapRowToStruct(t *testing.T) {
	db := GetConnection(t)
	user := &User{Name: "John Doe", Email: "john.doe@acme.com"}
	_, err := db.Exec("INSERT INTO users(name,email) values(?,?)", user.Name, user.Email)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT name as Name, email as Email, created as Created from users")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	users := []*User{}
	for rows.Next() {
		user := new(User)
		err = tools.MapRowToStruct(columns, rows, user, true)
		if err != nil {
			t.Fatal(err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	t.Log(users[0])
	if l := len(users); l != 1 {
		t.Fatalf("users length should be 1 , got %d", l)
	}
	if users[0].Email != user.Email {
		t.Fatalf("Error email should be '%s', got '%s'", user.Email, users[0].Email)
	}
}

func GetConnection(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: "./../testdata/migrations/development.sqlite3",
	}
	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
