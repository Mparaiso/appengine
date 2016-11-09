package query_test

import (
	"fmt"
	. "github.com/mparaiso/go-orm/query"
)

// A basic SELECT statement
func ExampleBuilder_BuildQuery() {

	result, values, err := Builder{From: []string{"articles"}}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)

	// Output:
	// <nil>
	// SELECT * FROM articles  ;
	// []
}

// A SELECT statement with columns and aliASes and a WHERE statement with parameters
func ExampleBuilder_BuildQuery_second() {

	result, values, err := Builder{
		Select: []string{"a.title AS Title", "a.created_at AS CreatedAt"},
		From:   []string{"articles a"},
		Where:  []string{"a.published", "=", "?"},
		Params: []interface{}{true},
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
func ExampleBuilder_BuildQuery_third() {

	result, values, err := Builder{
		Select: []string{"a.title AS Title"},
		From:   []string{"articles a"},
		Join:   []Join{{Table: "users u", On: "a.author_id = user.id"}},
		Where:  []string{"u.id", "=", "?"},
		Params: []interface{}{1},
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
func ExampleBuilder_BuildQuery_fourth() {
	results, values, err := Builder{
		Select:     []string{"u.id AS ID"},
		From:       []string{"users u"},
		Join:       []Join{{Table: "followers f", On: "f.followee_id = u.id"}},
		Aggregates: []Aggregate{{Type: COUNT, Column: "f.followee_id", As: "FollowerCount"}},
		GroupBy:    []string{"u.id"},
		Where:      []string{"u.id", "=", "?"},
		Params:     []interface{}{10},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(results)
	fmt.Println(values)

	// Output:
	// <nil>
	// SELECT COUNT(f.followee_id) AS FollowerCount, u.id AS ID FROM users u JOIN followers f ON f.followee_id = u.id WHERE u.id = ? GROUP BY u.id ;
	// [10]

}

// An example of a DELETE statement
func ExampleBuilder_BuildQuery_fifth() {
	result, values, err := Builder{
		Type:   DELETE,
		From:   []string{"articles"},
		Where:  []string{"articles.author_id", "=", "?"},
		Params: []interface{}{nil},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(values)

	// Output:
	// <nil>
	// DELETE FROM articles WHERE articles.author_id = ?;
	// [<nil>]
}

// An example of an UPDATE statement
func ExampleBuilder_BuildQuery_sixth(){
	query,values,err:=Builder{
		Update:"articles",
		Set:map[string]interface{}{
			"title":"the title",
		},
		Where:[]string{
			"articles.id","=","?",
		},
		Params:[]interface{}{1},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(query)
	fmt.Println(values)

	// Output:
	// <nil>
	// UPDATE articles SET title = ? WHERE articles.id = ?;
	// [the title 1]
}

//An example of an INSERT statement
func ExampleBuilder_BuildQuery_seventh(){
	query,values,err:=Builder{
		Insert:"articles",
		Set:map[string]interface{}{
			"title":"new title",
		},
	}.BuildQuery()

	fmt.Println(err)
	fmt.Println(query)
	fmt.Println(values)

	// Output:
	// <nil>
	// INSERT INTO articles(title) VALUES(?);
	// [new title]
}
