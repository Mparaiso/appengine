package orm_test

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"

	. "github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
	. "github.com/smartystreets/goconvey/convey"
)

func TestORM(t *testing.T) {

	Convey("Given a datamapper", t, func() {
		orm := NewORM(GetConnection(t))
		defer orm.Connection.Close()

		Convey("When an entity is registered", func() {
			err := orm.Register(new(User), new(Article))
			Reset(func() {
				repository, err := orm.GetRepository(new(User))
				So(err, ShouldBeNil)
				err = repository.DeleteAll()
				So(err, ShouldBeNil)
				repository, err = orm.GetRepository(new(Article))
				So(err, ShouldBeNil)
				err = repository.DeleteAll()
				So(err, ShouldBeNil)
			})
			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("The datamapper has a repository for the entity", func() {
				userRepository, err := orm.GetRepository(new(User))
				So(err, ShouldBeNil)
				So(userRepository, ShouldNotBeNil)
			})

			Convey("When an entity is persisted and flushed", func() {
				user := &User{Name: "John Doe", Email: "john.doe@acme.com"}
				orm.Persist(user)
				err := orm.Flush()
				Convey("Error should be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("The entity should have a valid ID", func() {
					So(user.ID, ShouldBeGreaterThan, 0)
				})

				Convey("Given a persisted entity", func() {
					Convey("When an entity is updated and flushed", func() {
						newName := "Marc Bolan"
						user.Name = newName
						orm.Persist(user)
						err := orm.Flush()

						Convey("Error should be nil", func() {
							So(err, ShouldBeNil)
						})

						Convey("The reloaded entity should have the correct modifications", func() {
							updatedUser := new(User)
							userRepository, err := orm.GetRepository(updatedUser)
							So(err, ShouldBeNil)
							userRepository.Find(user.ID, updatedUser)
							So(updatedUser.Name, ShouldEqual, user.Name)
						})
					})
				})
			})
			Convey("When multiple entities are persisted and flushed", func() {
				users := GetUserFixture()
				for _, user := range users {
					orm.Persist(user)
				}
				err := orm.Flush()

				Convey("Error should be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("Each entity has a valid ID", func() {
					for index, user := range users {
						So(user.ID, ShouldBeGreaterThan, 0)
						if index > 0 {
							So(user.ID, ShouldBeGreaterThan, users[index-1].ID)
						}
					}
				})

				Convey("Given an entity repository", func() {
					userRepository, err := orm.GetRepositoryByEntityName("User")
					So(err, ShouldBeNil)

					Convey("When an entity is fetched with repository.Find", func() {
						result := new(User)
						err = userRepository.Find(users[0].ID, result)

						Convey("There should be no error", func() {
							So(err, ShouldBeNil)
						})

						Convey("The result should have the right ID", func() {
							So(result.ID, ShouldBeGreaterThan, 0)
							So(result.ID, ShouldEqual, users[0].ID)
						})
					})

					Convey("When entities are fetched with repository.FindBy", func() {
						results := []*User{}
						err = userRepository.FindBy(
							Query{
								Where:  []string{"ID", "IN", "(", "?", ",", "?", ")"},
								Params: array{users[0].ID, users[1].ID}, OrderBy: map[string]Order{"ID": ASC},
							}, &results)
						So(err, ShouldBeNil)
						Convey("The result should have the right number of entities and the correct IDs", func() {
							So(len(results), ShouldEqual, 2)
							So(results[0].ID, ShouldEqual, users[0].ID)
							So(results[1].ID, ShouldEqual, users[1].ID)
						})
					})

					Convey("When entities are fetched with repository.All", func() {
						results := []*User{}
						err = userRepository.All(&results)
						Convey("There should be no error", func() {
							So(err, ShouldBeNil)
						})
						Convey("The result should have the right number of entities", func() {
							So(len(results), ShouldEqual, len(users))
						})
					})

				})

			})
		})
	})
}

func TestORMPersist(t *testing.T) {
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(User), new(Article))
	if err != nil {
		t.Fatal(err)
	}
	user := &User{Name: "John", Email: "john@acme.com"}
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("user.ID", user.ID)
	user.Name = "Jack"
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
}

func TestORMDestroy(t *testing.T) {
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(User), new(Article))
	if err != nil {
		t.Fatal(err)
	}
	user := &User{Name: "John", Email: "john@acme.com"}
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	orm.Destroy(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
}

type array []interface{}

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
			{
				StructField:  "Articles",
				Type:         OneToMany,
				TargetEntity: "Article",
				IndexBy:      "AuthorID",
				MappedBy:     "Author",
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
