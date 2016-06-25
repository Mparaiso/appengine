package query_test
import(
	.	"github.com/mparaiso/go-orm/query"
	"fmt"
)
// Example shows a basic SELECT statement
func Example(){
	result,values,err:=Query{From:[]string{"articles"}}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)
	// Output:
	// <nil>
	// SELECT * FROM articles ;
}
// Example_second show a SELECT statement with columns and aliases and a WHERE statement with parameters
func Example_second(){
	result,values,err  := Query{Type:SELECT,Select:[]string{"a.title as Title","a.created_at as CreatedAt"},
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
