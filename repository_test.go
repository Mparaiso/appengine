package orm_test

import "testing"
import . "github.com/mparaiso/go-orm"

func TestRepositoryFind(t *testing.T) {
	orm := GetORM(t)
	user := GetUserFixture()[0]
	articles := GetArticleFixture()
	orm.MustRegister(new(User), new(Article))
	orm.Persist(user).MustFlush()
	user.AddArticles(articles...)
	orm.Persist(user).MustFlush()

	userRepository, err := orm.GetRepository(user)

	if err != nil {
		t.Fatal(err)
	}
	u := new(User)
	err = userRepository.Find(user.ID, u)
	if err != nil {
		t.Fatal(err)
	}
	//t.Log(u)
	if l := len(u.Articles); l != 2 {
		t.Fatalf("Articles length should be 2, got %d", l)
	}
}

func TestRepositoryFindBy(t *testing.T) {
	user := &User{Name: "John Doe", Email: "john.doe@acme.com"}
	articles := []*Article{{Title: "First Article"}, {Title: "Second Article"}}
	orm := GetORM(t)
	orm.Persist(user).MustFlush()
	user.AddArticles(articles...)
	orm.Persist(user).MustFlush()
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

func TestResolveOneToOneSingle(t *testing.T) {
	// SetUp
	orm := GetORM(t)
	defer orm.Connection().Close()
	john := &User{Name: "John Doe", Email: "john.doe@acme.com"}
	johnInfo := &UserInfo{NiceName: "John", DisplayName: "J.Doe"}
	john.SetUserInfo(johnInfo)
	orm.Persist(john).MustFlush()
	repository, err := orm.GetRepositoryByEntityName("User")
	if err != nil {
		t.Fatal(err)
	}
	// Test
	result := new(User)
	err = repository.Find(john.ID, result)
	if err != nil {
		t.Fatal(err)
	}
	if result.UserInfo == nil {
		t.Fatal("user.UserInfo should not be nil")
	}
	if result.UserInfo.NiceName != johnInfo.NiceName {
		t.Fatalf("result.UserInfo.NiceName should be '%S' got '%s'", johnInfo.NiceName, result.UserInfo.NiceName)
	}
}

func TestResolveOneToOne(t *testing.T) {
	// SetUp
	orm := GetORM(t)
	users := []*User{
		{Name: "John Doe", Email: "john.doe@acme.com", UserInfo: &UserInfo{NiceName: "John D."}},
		{Name: "Jane Doe", Email: "jane.doe@acme.com", UserInfo: &UserInfo{NiceName: "Jane D."}},
	}
	err := orm.Persist(users[0], users[1]).Flush()
	if err != nil {
		t.Fatal(err)
	}
	repository, err := orm.GetRepositoryByEntityName("User")
	if err != nil {
		t.Fatal(err)
	}
	result := []*User{}
	err = repository.FindBy(Query{
		OrderBy: map[string]Order{"ID": ASC},
		Where:   []string{"ID", "IN", "(", "?", ",", "?", ")"},
		Params:  array{users[0].ID, users[1].ID},
	}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if l, m := len(result), len(users); l != m {
		t.Fatalf("Length should be %d, got %d", m, l)
	}
	if result[0].UserInfo == nil {
		t.Fatal("result[0].UserInfo should not be nil")
	}
	if e, a := result[0].UserInfo.NiceName, users[0].UserInfo.NiceName; e != a {
		t.Fatalf("NiceName should be %s, got %s", e, a)
	}
}
