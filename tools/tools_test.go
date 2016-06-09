package tools_test

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/go-orm/tools"
	"github.com/rubenv/sql-migrate"
)

import (
	"testing"
)

type User struct {
	Name  string
	Email string
	*UserInfo
}
type UserInfo struct {
	Url string
}

func Test(t *testing.T) {
	t.Log("test")
	db := GetConnection(t)
	user := &User{"John Doe", "john.doe@acme.com", nil}
	_, err := db.Exec("INSERT INTO users(name,email) values(?,?)", user.Name, user.Email)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT name as Name, email as Email from users")
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
		err = tools.MapRowToStruct(columns, rows, user)
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
		Dir: "./../test/testdata/migrations/development.sqlite3",
	}
	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
