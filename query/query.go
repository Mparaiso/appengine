package query

import (
	"fmt"
	"strings"
)

// Order is  an order
type Order string

const (
	// ASC in an ORDER BY statement
	ASC Order = "ASC"
	// DESC in an ORDER BY statement
	DESC Order = "DESC"
)

// Type is the statement type
type Type int8

const (
	// SELECT represents a SELECT statement
	SELECT Type = iota
	// INSERT represents an INSERT statement
	INSERT
	// DELETE represents a DELETE statement
	DELETE
	// UPDATE represents an UPDATE statement
	UPDATE
)

type aggregateType string

const (
	COUNT   aggregateType = "COUNT"
	AVERAGE aggregateType = "AVG"
	MIN     aggregateType = "MIN"
	MAX     aggregateType = "MAX"
	SUM     aggregateType = "SUM"
)

type Join struct {
	Table string
	On    string
}

// Aggregate describes an aggregate operation like COUNT, SUM, MIN, MAX, AVG
type Aggregate struct {
	Type aggregateType
	As   string
	// StructField is the struct field that should be populated with the result of the aggregate
	Column string
	// Separator used by GROUP_CONCAT
	Separator string
}

// Query implements QueryBuilder.
// It can be used to filter entities when they are fetched
// from the database.
type Builder struct {
	Type       Type
	From       []string
	Select     []string
	Update     string
	Delete     string
	Insert     string
	Where      []string
	Params     []interface{}
	OrderBy    map[string]Order
	Join       []Join
	Limit      int64
	Offset     int64
	Aggregates []Aggregate
	GroupBy    []string
	Set        map[string]interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *Builder {
	return &Builder{}
}

// GetType return the type of the query
func (query Builder) GetType() Type {
	return query.Type
}

// BuildQuery builds the querystring with placeholders , returns it with the parameters
// to replace the placeholders and an error.
func (query Builder) BuildQuery() (string, []interface{}, error) {
	if query.Update != "" {
		query.Type = UPDATE
	} else if query.Delete != "" {
		query.Type = DELETE
	} else if query.Insert != "" {
		query.Type = INSERT
	} else if len(query.Select) == 0 {
		query.Select = []string{"*"}
	}
	switch query.Type {

	case SELECT:
		values := []interface{}{}
		selectStatement, err := query.buildSelectStatement()
		if err != nil {
			return "", values, err
		}
		fromStatement := query.buildFromStatement()
		joinStatement := ""
		for _, join := range query.Join {
			joinStatement += "JOIN " + join.Table + " ON " + join.On + " "
		}

		whereStatement, values, err := query.buildWhereStatement()
		if err != nil {
			return "", values, err
		}
		groupByStatement := ""
		for i, group := range query.GroupBy {
			if i == 0 {
				groupByStatement += " GROUP BY"
			}
			groupByStatement += " " + group + " ,"
		}
		groupByStatement = strings.TrimRight(groupByStatement, " ,")
		orderByStatement, err := query.buildOrderByStatment()
		if err != nil {
			return "", values, err
		}
		limitStatement := query.buildLimitStatement()
		offsetStatement := query.buildOffsetStatement()

		q := []string{selectStatement, fromStatement, joinStatement, whereStatement, groupByStatement, orderByStatement, limitStatement, offsetStatement, " ;"}
		return strings.Join(q, ""), values, nil

	case UPDATE:
		updateStatement, values := query.buildUpdateStatement()
		whereStatement, v, err := query.buildWhereStatement()
		if err != nil {
			return "", values, err
		}
		values = append(values, v...)
		q := strings.Join([]string{updateStatement, whereStatement, ";"}, "")
		return q, values, nil

	case DELETE:

		deleteStatement := fmt.Sprintf("DELETE FROM %s ", strings.Join(query.From, ","))
		whereStatement, values, err := query.buildWhereStatement()
		if err != nil {
			return "", values, err
		}
		limitStatement := query.buildLimitStatement()
		query := deleteStatement + whereStatement + limitStatement + ";"
		return query, values, nil

	case INSERT:
		return query.buildInsertStatment()

	default:
		return "", []interface{}{}, fmt.Errorf("The query type %v is not supported.", query.Type)
	}

}

func (query Builder) buildInsertStatment() (string, []interface{}, error) {
	values := []interface{}{}
	columns := []string{}

	for key, value := range query.Set {
		columns = append(columns, key)
		values = append(values, value)
	}

	q := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s);",
		query.Insert,
		strings.Join(columns, ","),
		strings.Join(
			strings.Split(strings.Repeat("?", len(columns)), ""), ","))
	return q, values, nil
}

func (query Builder) buildUpdateStatement() (string, []interface{}) {
	setStatement := ""
	values := []interface{}{}
	fieldMap := query.Set
	for key, value := range fieldMap {
		setStatement = fmt.Sprintf("%s %s = ? ,", setStatement, key)
		values = append(values, value)
	}
	setStatement = strings.TrimLeft(strings.TrimSuffix(setStatement, " ,"), " ")
	updateStatement := fmt.Sprintf("UPDATE %s SET %s", query.Update, setStatement)
	return updateStatement, values
}

func (query Builder) buildLimitStatement() string {
	if query.Limit != 0 {
		return fmt.Sprintf(" LIMIT %d ", query.Limit)
	}
	return ""
}

func (query Builder) buildOffsetStatement() string {
	if query.Offset != 0 {
		return fmt.Sprintf(" OFFSET %d ", query.Offset)
	}
	return ""
}

func (query Builder) buildOrderByStatment() (string, error) {
	orderByStatement := ""
	if query.OrderBy != nil {
		for key, value := range query.OrderBy {
			if orderByStatement == "" {
				orderByStatement = fmt.Sprintf("%s %s", key, value)
			} else {
				orderByStatement = fmt.Sprintf("%s , %s %s ", orderByStatement, key, value)
			}
		}
		return " ORDER BY " + orderByStatement, nil
	}
	return "", nil
}

func (query Builder) buildSelectStatement() (string, error) {
	buildFromStatement := ""
	fieldListStatement := strings.Join(query.Select, ",")
	// Create aggregation statements ( like "COUNT(TABLE.COLUMN1) AS ALIAS" )
	if len(query.Aggregates) > 0 {
		for _, aggregate := range query.Aggregates {
			buildFromStatement = buildFromStatement + string(aggregate.Type) + "(" + aggregate.Column + ") "
			if aggregate.As != "" {
				buildFromStatement += "AS " + aggregate.As
			}
		}
		if fieldListStatement != "" {
			buildFromStatement = buildFromStatement + ", "
		}
	}
	return fmt.Sprintf("SELECT %s%s ", buildFromStatement, fieldListStatement), nil
}

func (query Builder) buildFromStatement() string {
	return fmt.Sprintf("FROM %s ", strings.Join(query.From, ","))
}

func (query Builder) buildWhereStatement() (string, []interface{}, error) {
	// values to be added as query parameters
	values := []interface{}{}
	if len(query.Where) > 0 {
		if len(query.Params) > 0 {
			values = append(values, query.Params...)
		}
		return "WHERE " + strings.Join(query.Where, " "), values, nil
	}
	return "", values, nil
}

// Returns the index of the needle in sliceOfString or
// -1 if the needle was not found.
func indexOfString(sliceOfString []string, needle string) int {
	for index, element := range sliceOfString {
		if element == needle {
			return index
		}
	}
	return -1
}
