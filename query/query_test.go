package query_test
import(
	.	"github.com/mparaiso/go-orm/query"
	"fmt"
)
// A basic SELECT statement
func ExampleQuery_BuildQuery(){
	result,values,err:=Query{From:[]string{"articles"}}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)
	
	// Output:
	// <nil>
	// SELECT * FROM articles  ;
	// []
}
// A SELECT statement with columns and aliases and a WHERE statement with parameters
func ExampleQuery_BuildQuery_second(){
	result,values,err  := Query{Select:[]string{"a.title as Title","a.created_at as CreatedAt"},
		From:[]string{"articles a"},
		Where:[]string{"a.published","=","?"},
		Params:[]interface{}{true},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)
	
	// Output:
	// <nil>
	// SELECT a.title as Title,a.created_at as CreatedAt FROM articles a WHERE a.published = ? ;
	// [true]
}

func ExampleQuery_BuildQuery_third(){
	result,values,err:=Query{Select:[]string{"a.title as Title"},
		From:[]string{"articles a"},
		Join:[]Join{{Table:"users u",On:"a.author_id = user.id"}},
		Where:[]string{"u.id","=","?"},
		Params:[]interface{}{1},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)

	// Output:
	// <nil>
	// SELECT a.title as Title FROM articles a JOIN users u ON a.author_id = user.id WHERE u.id = ? ;
	// [1]
}
