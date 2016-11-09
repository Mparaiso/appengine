package orm_test

import "testing"
import . "github.com/mparaiso/go-orm"

func TestQueryJoin(t *testing.T) {
	// set up
	orm := GetORM(t)
	articleRepository := orm.MustGetRepository(new(Article))
	userID := 1
	query := Query{Select: []string{"ID"}, Join: []Join{{"User"}}, Where: []string{"User.ID", "=", "?"}, Params: array{userID}}
	queryString, _, err := query.BuildQuery(articleRepository)
	if err != nil {
		t.Fatal(err)
	}
	expected := "SELECT articles.id AS ID FROM articles JOIN users ON users.id = articles.author_id WHERE users.id = ? ;"
	if actual := queryString; expected != actual {
		t.Fatalf(`
Expected =
'%s'
GOT = 
'%s'`, expected, actual)
	}
	//t.Log(userID, query, queryString, values)
}
