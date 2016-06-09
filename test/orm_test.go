package test_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
)

/*
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
	user.Name = "Jack"
	user.AddArticles([]*Article{{Title: "First Article Title"}, {Title: "Second Article Title"}}...)
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	ThenTestRepositoryFind(orm, t, user.ID)
}

func ThenTestRepositoryFind(orm *ORM, t *testing.T, userID int64) {
	userRepository, err := orm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	user := new(User)
	err = userRepository.Find(userID, user)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(user.Articles); l != 2 {
		t.Fatalf("Articles length should be 2, got %d", l)
	}
}

func TestRepositoryFindBy(t *testing.T) {
	user := &User{Name: "John Doe", Email: "john.doe@acme.com"}
	articles := []*Article{{Title: "First Article"}, {Title: "Second Article"}}
	orm := NewORM(GetConnection(t))
	orm.Register(new(User), new(Article))
	err := orm.Persist(user).Flush()
	if err != nil {
		t.Fatal(err)
	}
	user.AddArticles(articles...)
	err = orm.Persist(user).Flush()
	if err != nil {
		t.Fatal(err)
	}
	userRepository, err := orm.GetRepository(user)
	if err != nil {
		t.Fatal(err)
	}
	result := []*User{}
	err = userRepository.FindBy(Query{Where: []string{"ID", "=", "?"}, Params: []interface{}{user.ID}}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(user.Articles); l != 2 {
		t.Fatalf("length should b 2, got %d", l)
	}
	ThenTestRepositoryAll(orm, t)
}

func ThenTestRepositoryAll(orm *ORM, t *testing.T) {
	userRepository, err := orm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	result := []*User{}
	err = userRepository.All(&result)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 1 {
		t.Fatalf("Len Should be 1, got %d", l)
	}

	ThenTestRepositoryFindOneBy(userRepository, t, result[0])
}

func ThenTestRepositoryFindOneBy(repository *Repository, t *testing.T, user *User) {
	result := new(User)
	err := repository.FindOneBy(Query{Where: []string{"ID", "=", "?"}, Params: []interface{}{user.ID}}, result)
	if err != nil {
		t.Fatal(err)
	}
	if id := result.ID; id != user.ID {
		t.Fatalf("The ID of the user should be %d, got %d", user.ID, result.ID)
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
	t.Log("Destroy entity that has a one to many relationship with cascade Remove of owned entities")
	user = GetUserFixture()[1]
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	userID := user.ID
	articles := GetArticleFixture()
	user.AddArticles(articles[:2]...)
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
	articleRepository, err := orm.GetRepository(new(Article))
	if err != nil {
		t.Fatal(err)
	}
	result := []*Article{}
	articleRepository.FindBy(Query{Where: []string{"AuthorID", "=", "?"}, Params: array{userID}}, &result)
	if l := len(result); l != 0 {
		t.Fatalf("Length should be 0, got %d", l)
	}
}
*/
func TestRegressionBug(t *testingT) {
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(Article), new(User), new(UserInfo))
	userRepository, err := orm.GetRepository(new(User))
	query := Query{}
}
func TestRegressionBug2(t *testing.T) {
	t.Skip()
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(Article), new(User), new(UserInfo))
	if err != nil {
		t.Fatal(err)
	}
	userFixtures := GetUserFixture()
	user := userFixtures[0]
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	fetchedUser := new(User)
	userRepository, err := orm.GetRepository(user)
	if err != nil {
		t.Fatal(err)
	}
	err = userRepository.Find(user.ID, fetchedUser)
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
