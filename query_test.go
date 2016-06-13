package orm_test

import "testing"

func TestQueryJoin(t *testing.T) {
	// set up
	orm := GetORM(t)
	articleRepository := orm.MustGetRepository(new(Article))
	userID := 1
	query := Query{Select: []string{"ID"}, Join: []string{"User"}, Where: []string{"User.ID", "=", "?"}, Params: array{1}}
	queryString, values, err := query.BuildQuery(articleRepository)
	if err != nil {
		t.Fatal(err)
	}
}
