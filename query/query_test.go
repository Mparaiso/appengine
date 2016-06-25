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
// A SELECT statement with columns and aliASes and a WHERE statement with parameters
func ExampleQuery_BuildQuery_second(){

	result,values,err  := Query{
		Select:[]string{"a.title AS Title","a.created_at AS CreatedAt"},
		From:[]string{"articles a"},
		Where:[]string{"a.published","=","?"},
		Params:[]interface{}{true},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)
	
	// Output:
	// <nil>
	// SELECT a.title AS Title,a.created_at AS CreatedAt FROM articles a WHERE a.published = ? ;
	// [true]
}

// A SELECT statement WITH A JOIN statement
func ExampleQuery_BuildQuery_third(){

	result,values,err:=Query{
		Select:[]string{"a.title AS Title"},
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
	// SELECT a.title AS Title FROM articles a JOIN users u ON a.author_id = user.id WHERE u.id = ? ;
	// [1]
}

// A select statement with aggregation
func ExampleQuery_BuildQuery_selection_with_join_and_group_by(){
	results,values,err:=Query{
		Select:[]string{"u.id AS ID"},
		From:[]string{"users u"},
		Join:[]Join{{Table:"followers f",On:"f.followee_id = u.id"}},
		Aggregates:[]Aggregate{{Type:COUNT,Column:"f.followee_id",As:"FollowerCount"}},
		GroupBy:[]string{"u.id"},
		Where:[]string{"u.id","=","?"},
		Params:[]interface{}{10},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(results)
	fmt.Println(values)

	// Output:
	// <nil>
	// SELECT COUNT(f.followee_id) AS FollowerCount, u.id AS ID FROM users u JOIN followers f ON f.followee_id = u.id WHERE u.id = ? GROUP BY u.id ;
	// [10]

}
