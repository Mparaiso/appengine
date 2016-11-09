package orm_test

import (
//	"fmt"
	"testing"

//	. "github.com/mparaiso/go-orm/orm"
)

type array []interface{}

func TestQueryBuilder_BuildQuery_select(t *testing.T) {}

func ExampleQueryBuilder_BuildQuery_select() {
	// query, values, err := QueryBuilder{
	// 	From:   []string{"Article"},
	// 	Where:  []string{"Article.ID", "=", "?"},
	// 	Params: array{1},
	// }.BuildQuery(nil)

	// fmt.Println(err)
	// fmt.Println(query)
	// fmt.Println(values)

	// Output
	// <nil>
	// SELECT FROM articles WHERE articles.id = ? ;
	// [1]
}
