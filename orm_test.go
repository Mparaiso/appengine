package orm_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/mparaiso/go-orm"
)

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

func TestRegressionBug2(t *testing.T) {

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
