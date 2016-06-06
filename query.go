package orm

import (
	"fmt"
	"strings"
)

type Order string

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

type QueryType int8

const (
	SELECT QueryType = iota
	DELETE
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

type Aggregate struct {
	Type aggregateType
	On   string
	// StructField is the struct field that should be populated with the result of the aggregate
	StructField string
	Separator   string
}

// Query implements QueryBuilder.
// It can be used to filter entities when they are fetched
// from the database.
type Query struct {
	Type       QueryType
	Select     []string
	Where      []string
	Params     []interface{}
	OrderBy    map[string]Order
	Join       []string
	Limit      int64
	Offset     int64
	Aggregates []Aggregate
}

func (query Query) BuildQuery(repository *Repository) (string, []interface{}, error) {

	switch query.Type {
	case SELECT:
		values := []interface{}{}
		selectStatement, err := query.BuildSelectStatement(repository)
		if err != nil {
			return "", values, err
		}
		fromStatement := query.BuildFromStatement(repository)
		whereStatement, v, err := query.BuildWhereStatement(repository)
		if err != nil {
			return "", v, err
		}
		values = append(values, v...)
		orderByStatement := query.BuildOrderByStatment(repository)
		limitStatement := query.BuildLimitStatement(repository)
		offsetStatement := query.BuildOffsetStatement(repository)
		q := []string{selectStatement, fromStatement, whereStatement, orderByStatement, limitStatement, offsetStatement, " ;"}
		return strings.Join(q, ""), values, nil
	default:
		return "", []interface{}{}, fmt.Errorf("The query type %v is not supported.", query.Type)
	}

}

func (query Query) BuildLimitStatement(repository *Repository) string {
	if query.Limit != 0 {
		return fmt.Sprintf(" LIMIT %d ", query.Limit)
	}
	return ""
}
func (query Query) BuildOffsetStatement(repository *Repository) string {
	if query.Offset != 0 {
		return fmt.Sprintf(" OFFSET %d ", query.Offset)
	}
	return ""
}

func (query Query) BuildOrderByStatment(repository *Repository) (string, []interface{}, error) {
	orderByStatement := ""
	if query.OrderBy != nil {
		metadata := repository.ORM.metadatas[repository.Type]
		for key, value := range query.OrderBy {

			columnName, ok := metadata.FindColumnNameForField(key)
			if !ok {
				return "", nil, fmt.Errorf("No column found for field %s in OrderBy Query Part.", columnName)
			}
			if orderByStatement == "" {
				orderByStatement = fmt.Sprintf("%s %s", strings.ToLower(columnName), value)
			} else {
				orderByStatement = fmt.Sprintf("%s , %s %s ", orderByStatement, strings.ToLower(columnName), value)
			}
		}
		return " ORDER BY " + orderByStatement, []interface{}{}, nil
	}
	return "", []interface{}{}, nil
}

func (query Query) BuildSelectStatement(repository *Repository) (string, error) {
	metadata := repository.ORM.GetTypeMetadata(repository.Type)
	aggregateListStatement := ""
	// Get columns to be returned ( like " table.column1 AS structField1 , table.column2 AS structField2 " )
	fieldListStatement, err := buildSelectFieldListFromColumnMetadata(metadata, query.Select...)
	if err != nil {
		return "", err
	}
	// Create aggregation statements ( like "COUNT(TABLE.COLUMN1) AS ALIAS" )
	if len(query.Aggregates) > 0 {
		for _, aggregate := range query.Aggregates {
			columnName, ok := metadata.FindColumnNameForField(aggregate.On)
			if !ok {
				return "", fmt.Errorf("StructField '%s' Not Found on aggregate %v .", aggregate.On, aggregate)
			}
			aggregateListStatement = aggregateListStatement + " " + string(aggregate.Type) + "(" + repository.TableName + "." + columnName + ") AS " + aggregate.StructField
		}
		if fieldListStatement != "" {
			aggregateListStatement = aggregateListStatement + ", "
		}
	}
	return fmt.Sprintf("SELECT %s%s ", aggregateListStatement, fieldListStatement), nil
}

func (query Query) BuildFromStatement(repository *Repository) string {
	return fmt.Sprintf(" FROM %s ", repository.TableName)
}

func (query Query) BuildWhereStatement(repository *Repository) (string, []interface{}, error) {
	values := []interface{}{}
	metadata := repository.ORM.metadatas[repository.Type]
	if query.Where != nil {
		if query.Params != nil {
			values = append(values, query.Params...)
		}
		paramNumber := 0
		for index, token := range query.Where {
			switch token {
			case "?":
				paramNumber = paramNumber + 1
			case "=", "<", "<=", ">", ">=", "<>", "!=", "IN", "NOT IN", "NOT LIKE", "LIKE":
				if index == 0 {
					return "", nil, fmt.Errorf("Unexpected token %s at index %d in Query.Where", token, index)
				}
				fieldName := query.Where[index-1]
				columnName, ok := metadata.FindColumnNameForField(fieldName)
				if !ok {
					return "", nil, fmt.Errorf("No column found for field %s in Where Query Part.", fieldName)
				}
				query.Where[index-1] = metadata.Table.Name + "." + strings.ToLower(columnName)
			}

		}
		if paramNumber != len(values) {
			return "", nil, fmt.Errorf("Not enough ? placeholders for params %v in %s ", values, query.Where)
		}
		return " WHERE " + strings.Join(query.Where, " "), values, nil
	}
	return "", values, nil
}

// buildSelectFieldListFromColumnMetadata uses  metadata to output
// a string to be used in a SELECT Statement : "TABLENAME.COLUMN1 AS COLUMN1, TABLENAME.COLUMN2 AS COLUMN2, ..."
// or returns an error if needed. The goal is to easily match struct field names in a query without having to use struct field tags
// by using a native SQL feature of aliasing column names.
// It is possible to pass optional struct field names to filter what data should be included in the result.
func buildSelectFieldListFromColumnMetadata(metadata Metadata, fieldNameSelector ...string) (string, error) {
	fields := ""
	for _, column := range metadata.Columns {
		if len(fieldNameSelector) == 1 && fieldNameSelector[0] == "" {
			return "", nil
		}
		if (len(fieldNameSelector) > 0 && indexOfString(fieldNameSelector, column.StructField) >= 0) || len(fieldNameSelector) == 0 {
			if column.Name == "" {
				fields = fields + metadata.Table.Name + "." + strings.ToLower(column.StructField) + " AS " + strings.ToLower(column.StructField) + ","
			} else {
				fields = fields + metadata.Table.Name + "." + column.Name + " AS " + strings.ToLower(column.StructField) + ","
			}
		}
	}
	// we remove the last comma in the string

	return string(fields[:len(fields)-1]), nil
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
