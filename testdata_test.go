package orm_test

import (
	"database/sql"
	"testing"
	"time"

	. "github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
	"golang.org/x/crypto/bcrypt"
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

func (Article) ProvideMetadata() Metadata {
	return Metadata{
		Entity: "Article",
		Table:  Table{Name: "articles"},
		Columns: []Column{
			{ID: true, Field: "ID"},
			{Field: "Title"},
			{Field: "Content"},
			{Field: "Created"},
			{Field: "Updated"},
			{Field: "AuthorID", Name: "author_id"},
		},
	}
}

func (a *Article) BeforeCreate() (err error) {
	a.Created = time.Now()
	a.Updated = time.Now()
	return
}

func (a *Article) BeforeUpdate() (err error) {
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
	UserInfo       *UserInfo
}

func (user *User) SetUserInfo(userInfo *UserInfo) {
	user.UserInfo = userInfo
	userInfo.User = user
	userInfo.UserID = user.ID
}
func (User) ProvideMetadata() Metadata {
	return Metadata{
		Entity: "User",
		Table:  Table{Name: "users"},
		Columns: []Column{
			{ID: true, Field: "ID"},
			{Field: "Email"},
			{Field: "Name"},
			{Field: "Created"},
			{Field: "Updated"},
			{Field: "PasswordDigest", Name: "password_digest"},
		},
		Relations: []Relation{
			{
				Field:        "Articles",
				Type:         OneToMany,
				TargetEntity: "Article",
				IndexedBy:    "AuthorID",
				MappedBy:     "Author",
				Fetch:        Eager,
				Cascade:      Persist | Remove,
			},
			{
				Field:        "UserInfo",
				Type:         OneToOne,
				TargetEntity: "UserInfo",
				IndexedBy:    "UserID",
				Cascade:      Persist | Remove,
				Fetch:        Eager,
			},
		},
	}
}

func (user *User) BeforeCreate() (err error) {
	user.Created = time.Now()
	user.Updated = time.Now()
	return
}

func (user *User) BeforeUpdate() (err error) {
	user.Updated = time.Now()
	return
}

func (user *User) AddArticles(articles ...*Article) {
	if user.Articles == nil {
		user.Articles = []*Article{}
	}
	for _, article := range articles {
		user.Articles = append(user.Articles, article)
		article.AuthorID = user.ID
	}
}

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

// UserInfo complements the user entity
type UserInfo struct {
	ID            int64
	NiceName      string
	URL           string
	Registered    time.Time
	ActivationKey string
	Status        int8
	DisplayName   string
	UserID        int64
	User          *User
}

func (UserInfo) ProvideMetadata() Metadata {
	return Metadata{
		Table:  Table{Name: "userinfos"},
		Entity: "UserInfo",
		Columns: []Column{
			{Field: "ID", ID: true},
			{Field: "NiceName"},
			{Field: "Registered"},
			{Field: "ActivationKey", Name: "activation_key"},
			{Field: "Status"},
			{Field: "DisplayName", Name: "display_name"},
			{Field: "UserID", Name: "user_id"},
		},
	}
}

// NotEntity is not a valid entity
type NotEntity struct{}

func GetUserFixture() []*User {
	return []*User{
		{Name: "john doe", Email: "john.doe@acme.com"},
		{Name: "jane doe", Email: "jane.doe@acme.com"},
		{Name: "bill doe", Email: "bill.doe@acme.com"},
		{Name: "suzy doe", Email: "suzy.doe@acme.com"},
	}
}

func GetArticleFixture() []*Article {
	return []*Article{
		{Title: "First Article Content", Content: "First Article Content"},
		{Title: "Second Article Content", Content: "Second Article Content"},
	}
}

func GetConnection(t *testing.T) *Connection {
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

	return NewConnectionWithOptions("sqlite3", db, &ConnectionOptions{Logger: t})
}

func GetORM(t *testing.T) *ORM {
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(User), new(UserInfo), new(Article))
	if err != nil {
		t.Fatal(err)
	}
	return orm
}
