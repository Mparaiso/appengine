package datamapper_test

import (
	"database/sql"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	. "github.com/mparaiso/datamapper"
	"github.com/rubenv/sql-migrate"
)

// Article is an article
type Article struct {
	ID       int64
	Title    string
	Content  string
	Created  time.Time
	Updated  time.Time
	Author   *User
	AuthorID int64
}

func (Article) DataMapperMetaData() Metadata {
	return Metadata{
		Entity: "Article",
		Table:  Table{Name: "articles"},
		Columns: []Column{
			{ID: true, StructField: "ID"},
			{StructField: "Title"},
			{StructField: "Content"},
			{StructField: "Created"},
			{StructField: "Updated"},
			{StructField: "AuthorID", Name: "author_id"},
		},
	}
}

func (a *Article) BeforeCreate() (err error) {
	a.Created = time.Now()
	return
}

func (a *Article) BeforeSave() (err error) {
	a.Updated = time.Now()
	return
}

// User is a user
type User struct {
	ID             int64
	Name           string
	Email          string
	Created        time.Time
	Updated        time.Time
	PasswordDigest string
	Articles       []*Article
}

func (User) DataMapperMetaData() Metadata {
	return Metadata{
		Entity: "User",
		Table:  Table{Name: "users"},
		Columns: []Column{
			{ID: true, StructField: "ID"},
			{StructField: "Email"},
			{StructField: "Name"},
			{StructField: "Created"},
			{StructField: "Updated"},
			{StructField: "PasswordDigest", Name: "password_digest"},
		},
		Relations: []Relation{
			{StructField: "Articles", Type: OneToMany, TargetEntity: "Article", IndexBy: "AuthorID", MappedBy: "Author", Fetch: Eager},
		},
	}
}

func (user *User) BeforeCreate() (err error) {
	user.Created = time.Now()
	return
}

func (user *User) BeforeSave() (err error) {
	user.Updated = time.Now()
	return
}

// BeforeSave does some work before saving
// see http://stackoverflow.com/questions/23259586/bcrypt-password-hashing-in-golang-compatible-with-node-js
func (user *User) GenerateSecurePassword(password string) error {
	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(passwordDigest)
	return nil
}

// Authenticate return an error if the password and PasswordDigest do not match
func (user User) Authenticate(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))
}

func userFixture() []*User {
	return []*User{
		{Name: "john doe", Email: "john.doe@acme.com"},
		{Name: "jane doe", Email: "jane.doe@acme.com"},
		{Name: "bill doe", Email: "bill.doe@acme.com"},
		{Name: "suzy doe", Email: "suzy.doe@acme.com"},
	}
}

// before initialize in memory database
func before(t *testing.T) *DataMapper {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: "testdata/migrations/development.sqlite3",
	}
	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}

	connection := NewConnectionWithOptions("sqlite3", db, &ConnectionOptions{Logger: t})
	dm := NewDataMapper(connection)
	err = dm.Register(new(User), new(Article))
	if err != nil {
		t.Fatal(err)
	}
	return dm
}
