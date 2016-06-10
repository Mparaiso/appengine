package orm_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
)

type AppUser struct {
	Name, Email string
	*UserInfos
}

type UserInfos struct {
}

func TestConnectionGet(t *testing.T) {
	connection := orm.NewConnection("sqlite3", GetDB(t))
	err := LoadFixtures(connection)
	if err != nil {
		t.Fatal(err)
	}
	user := new(AppUser)
	err = connection.Get(user, "SELECT name as Name,email as Email from users ;")
	if err != nil {
		t.Fatal(err)
	}
	if expected, actual := "John Doe", user.Name; expected != actual {
		t.Fatalf("Expected '%s', got '%s'", expected, actual)
	}
	notAPointerUser := AppUser{}
	err = connection.Get(notAPointerUser, "SELECT * from users ;")

	if e, ok := err.(orm.NotAPointerError); !ok {
		t.Fatalf("Error should be orm.NotAPointerError when passing non-pointer to connection.Get, got %#v", e)
	}

}

func TestConnectionSelect(t *testing.T) {
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
		//t.Log(result.LastInsertId())
	}

	// test query
	users := []*AppUser{}
	err = connection.Select(&users, "SELECT users.name as Name, users.email as Email from users ORDER BY users.id ASC ;")
	if err != nil {
		t.Fatal(err)
	} else {
		//t.Logf("%#v", users)
	}
	if expected, name := "john doe", users[0].Name; name != expected {
		t.Fatalf("users[0].Name should be '%s', got '%s'", expected, name)
	}
}

func LoadFixtures(connection *orm.Connection) error {
	for _, user := range []AppUser{
		{Name: "John Doe", Email: "john.doe@acme.com"},
		{Name: "Jane Doe", Email: "jane.doe@acme.com"},
		{Name: "Jack Doe", Email: "jack.doe@acme.com"},
	} {
		_, err := connection.Exec("INSERT INTO users(name,email) values(?,?);", user.Name, user.Email)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: "./testdata/migrations/development.sqlite3",
	}
	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
