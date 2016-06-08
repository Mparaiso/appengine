package test_test

import (
	"time"

	. "github.com/mparaiso/go-orm"
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
	UserInfoID     int64
	UserInfo       *UserInfo
}

func (User) ProvideMetadata() Metadata {
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
			{StructField: "UserInfoID", Name: "user_infos_id"},
		},
		Relations: []Relation{
			{
				StructField:  "Articles",
				Type:         OneToMany,
				TargetEntity: "Article",
				IndexBy:      "AuthorID",
				MappedBy:     "Author",
				Fetch:        Eager,
				Cascade:      Persist | Remove,
			},
			/*{
				StructField:  "UserInfo",
				Type:         OneToOne,
				TargetEntity: "UserInfo",
				JoinColumn:   JoinColumn{StructField: "UserInfoID", ReferencedStructField: "ID"},
			},*/
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

type UserInfo struct {
	ID            int64
	UserID        int64
	User          *User
	NiceName      string
	Url           string
	Registered    time.Time
	ActivationKey string
	Status        int
	DisplayName   string
}

func (UserInfo) ProvideMetadata() Metadata {
	return Metadata{
		Entity: "UserInfo",
		Table:  Table{Name: "user_infos"},
		Columns: []Column{
			{StructField: "ID", ID: true, Name: "id"},
			{StructField: "UserID", Name: "user_id"},
			{StructField: "NiceName", Name: "nicename"},
			{StructField: "Url", Name: "url"},
			{StructField: "Registered", Name: "registered"},
			{StructField: "ActivationKey", Name: "activation_key"},
			{StructField: "Status", Name: "status"},
			{StructField: "DisplayName", Name: "display_name"},
		},
	}
}

// NotEntity is not a valid entity
type NotEntity struct{}
