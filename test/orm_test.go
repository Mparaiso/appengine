package test_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
	. "github.com/smartystreets/goconvey/convey"
)

func TestORM(t *testing.T) {

	Convey("Given an ORM", t, func() {
		orm := NewORM(GetConnection(t))
		defer orm.Connection.Close()

		Convey("Given multiple entities registered in the ORM", func() {
			err := orm.Register(new(User), new(Article))
			userRepository, err := orm.GetRepository(new(User))
			So(err, ShouldBeNil)
			//articleRepository, err := orm.GetRepository(new(Article))
			//So(err, ShouldBeNil)
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
				err = orm.Flush()
				So(err, ShouldBeNil)
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
							userRepository.Find(user.ID, updatedUser)
							So(updatedUser.Name, ShouldEqual, user.Name)
						})
					})

					Convey("When an entity is destroyed", func() {
						id := user.ID
						orm.Remove(user)
						err = orm.Flush()
						So(err, ShouldBeNil)

						Convey("It shouldn't exist in the database as a record.", func() {
							u := new(User)
							err = userRepository.Find(id, u)
							So(err, ShouldEqual, sql.ErrNoRows)
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

			Convey("Given 2 entities with a OneToMany relationship", func() {
				user := &User{Name: "Jack Doe", Email: "jack.doe@acme.com"}
				orm.Persist(user)
				err = orm.Flush()
				So(err, ShouldBeNil)
				articles := []*Article{
					{Title: "First Article Title", Content: "First Article Content"},
					{Title: "Second Article Title", Content: "Second Article Content"},
				}
				user.AddArticles(articles...)
				//err = orm.Persist(articles[0], articles[1]).Flush()
				Convey("Given that the owning side relationship cascades on persist", func() {
					orm.Persist(user).Flush()
					So(err, ShouldBeNil)
					Convey("When the owning entity is fetched with repository.Find", func() {
						fetchedUser := new(User)
						err = userRepository.Find(user.ID, fetchedUser)
						So(err, ShouldBeNil)
						Convey("The owning entity should have many related entities", func() {
							So(fetchedUser.Articles, ShouldNotBeNil)
							So(len(fetchedUser.Articles), ShouldEqual, 2)
						})
					})
					Convey("When the owning entity is fetched with repository.FindBy", func() {
						fetchedUsers := []*User{}
						err = userRepository.FindBy(Query{Where: []string{"ID", "=", "?"}, Params: []interface{}{user.ID}}, &fetchedUsers)
						So(err, ShouldBeNil)
						So(len(fetchedUsers), ShouldEqual, 1)
						Convey("The owning entity should have many related entities", func() {
							So(fetchedUsers[0].Articles, ShouldNotBeNil)
							So(len(fetchedUsers[0].Articles), ShouldEqual, 2)
						})
					})
				})

				Convey("Given that the ownind side relationship cascades on delete", func() {

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
	orm.Remove(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
}

type array []interface{}

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

	return NewConnectionWithOptions("sqlite3", db, &ConnectionOptions{})
}
