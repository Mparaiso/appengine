package datamapper_test

import (
	"fmt"
	"testing"

	_ "github.com/amattn/go-sqlite3"
	. "github.com/mparaiso/datamapper"
)

type hash map[string]interface{}
type array []interface{}

func TestAll(t *testing.T) {
	dm := before(t)
	userRepository, err := dm.GetRepository(&User{})
	if err != nil {
		t.Fatal(err)
	}
	user := &User{Name: "John Doe", Email: "john.doe@acme.com"}
	err = userRepository.Save(user)
	users := []*User{}
	err = userRepository.All(&users)
	// t.Log("users length : ", len(users))
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatalf("len(users) should be 1, got %d", len(users))
	}
}

func TestFind(t *testing.T) {
	dm := before(t)
	defer dm.Connection.Close()
	userRepository, err := dm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	user := &User{Name: "John Doe", Email: "john.doe@acme.com", PasswordDigest: "password"}

	err = userRepository.Save(user)
	if err != nil {
		t.Fatal(err, "Error saving a user", user)
	}
	t.Logf("%#v", user)
	fetchedUser := new(User)
	err = userRepository.Find(user.ID, fetchedUser)
	if err != nil {
		t.Fatal(err, "Error find a user with id", user.ID)
	}
	if fetchedUser.PasswordDigest != "password" {
		t.Fatal("Wrong password", fetchedUser.PasswordDigest)
	}
}

func TestFindBy(t *testing.T) {
	dm := before(t)
	defer dm.Connection.Close()
	userRepository, err := dm.GetRepository(&User{})
	if err != nil {
		t.Fatal(err)
	}
	userName := "John Doe"
	user := &User{Name: userName, Email: "john.doe@acme.com"}

	err = userRepository.Save(user)
	if err != nil {
		t.Fatal(err)
	}
	err = userRepository.Save(&User{Name: "Jane Doe", Email: "jane.doe@acme.com"})
	if err != nil {
		t.Fatal(err)
	}
	if id := user.ID; id == 0 {
		t.Fatal("user.ID should be >0, got", id)
	}
	candidates := []*User{}
	err = userRepository.FindBy(Query{Where: []string{"Name", "=", "?"}, Params: array{userName}}, &candidates)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(candidates); l != 1 {
		t.Fatal(candidates, "length should be 1, got", l)
	}
}

func TestFindBy2(t *testing.T) {
	dm := before(t)
	defer dm.Connection.Close()
	repository, err := dm.GetRepositoryByTableName("users")
	if err != nil {
		t.Fatal(err)
	}
	users := userFixture()
	for _, user := range users {
		err = repository.Save(user)
		if err != nil {
			t.Fatal(err)
		}
	}
	result := []*User{}
	err = repository.FindBy(Query{
		Where:  []string{"Name", "=", "?", "AND", "Email", "=", "?"},
		Params: array{users[0].Name, users[0].Email},
	}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("len(result) should be 1, got %d.", len(result))
	}
}

func TestFindBy3(t *testing.T) {
	dm := before(t)
	defer dm.Connection.Close()
	users := userFixture()
	userRepository, err := dm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range users {
		err := userRepository.Save(user)
		if err != nil {
			t.Fatal(err)
		}
	}
	result := []*User{}
	err = userRepository.FindBy(Query{Where: []string{"ID", "IN", "(", "?", ",", "?", ")"}, Params: array{2, 3}}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) should be 2, got %d", len(result))
	}
}

func TestCount(t *testing.T) {
	dm := before(t)
	repository, err := dm.GetRepository(&User{})
	for i := 0; i < 3; i++ {
		err := repository.Save(&User{Name: fmt.Sprintf("user%d", i), Email: fmt.Sprintf("user%d@acme.com", i)})
		if err != nil {
			t.Fatal(i, err)
		}
	}
	count, err := repository.Count(Query{})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("count should be 3, got %d .", count)
	}
	count, err = repository.Count(Query{Where: []string{"Name", "=", "?"}, Params: array{"user1"}})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count should be 1, got %d", count)
	}

}
