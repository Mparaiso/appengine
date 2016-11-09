package orm_test

import (
	"database/sql"
	"flag"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

func TestConnectionSelectMap(t *testing.T) {
	connection := GetConnection(t)
	defer connection.Close()
	if err := LoadFixtures(connection); err != nil {
		t.Fatal(err)
	}
	var result []map[string]interface{}
	err := connection.SelectMap(&result, "SELECT * FROM users ORDER BY ID")
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 3 {
		t.Fatalf("length should be 3, got %d", l)
	}
	if created, ok := result[0]["created"].(time.Time); !ok {
		t.Fatalf("created should be time.Time, got %#v", created)
	}
}

func TestConnectionSelectSlice(t *testing.T) {
	connection := GetConnection(t)
	defer connection.Close()
	if err := LoadFixtures(connection); err != nil {
		t.Fatal(err)
	}
	var result [][]interface{}
	err := connection.SelectSlice(&result, "SELECT id,name,created FROM users ORDER BY ID")
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 3 {
		t.Fatalf("length should be 3, got %d", l)
	}
	if created, ok := result[0][2].(time.Time); !ok {
		t.Fatalf("created should be time.Time, got %#v", created)
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
	driver, datasource, migrationDirectory := "sqlite3", ":memory:", "./testdata/migrations/development.sqlite3"
	arguments := flag.Args()
	for _, argument := range arguments {
		switch argument {
		case "mysql":
			// go test ./... -v -run ConnectionGet -args mysql
			// https://github.com/go-sql-driver/mysql#examples
			driver, datasource, migrationDirectory = "mysql", "user@/test?parseTime=true", "./testdata/migrations/test.mysql"
		}
	}
	t.Log("Using driver ", driver)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationDirectory,
	}
	_, err = migrate.Exec(db, driver, migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func DropDB(t *testing.T) *sql.DB {
	driver, datasource, migrationDirectory := "sqlite3", ":memory:", "./testdata/migrations/development.sqlite3"
	arguments := flag.Args()
	for _, argument := range arguments {
		switch argument {
		case "mysql":
			// go test ./... -v -run ConnectionGet -args mysql
			driver, datasource, migrationDirectory = "mysql", "user@/test?parseTime=true", "./testdata/migrations/test.mysql"

		default:
			return nil
		}
	}
	t.Log("Using driver ", driver)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationDirectory,
	}
	_, err = migrate.Exec(db, driver, migrations, migrate.Down)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestMain(m *testing.M) {
	code := m.Run()
	DropDB(new(testing.T))
	os.Exit(code)

}
