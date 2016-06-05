package datamapper

import (
	"fmt"
	"strings"
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
	Select     []string
	Where      []string
	Params     []interface{}
	OrderBy    map[string]string
	Join       []string
	Limit      int64
	Offset     int64
	Aggregates []Aggregate
}

func (query Query) AcceptRepository(repository *Repository) (string, []interface{}, error) {
	values := []interface{}{}
	metadata := repository.DM.Metadatas[repository.Type]
	selectStatement := ""
	fromStatement := ""
	whereStatement := ""
	orderByStatement := ""
	limitStatement := ""
	offsetStatement := ""
	aggregateListStatement := ""

	// Get columns to be returned ( like " table.column1 AS structField1 , table.column2 AS structField2 " )
	fieldListStatement, err := buildSelectFieldListFromColumnMetadata(metadata, query.Select...)
	if err != nil {
		return "", values, err
	}
	// Create aggregation statements ( like "COUNT(TABLE.COLUMN1) AS ALIAS" )
	if len(query.Aggregates) > 0 {
		for _, aggregate := range query.Aggregates {
			columnName, ok := metadata.FindColumnNameForField(aggregate.On)
			if !ok {
				return "", values, fmt.Errorf("StructField '%s' Not Found on aggregate %v .", aggregate.On, aggregate)
			}
			aggregateListStatement = aggregateListStatement + " " + string(aggregate.Type) + "(" + repository.TableName + "." + columnName + ") AS " + aggregate.StructField
		}
		if fieldListStatement != "" {
			aggregateListStatement = aggregateListStatement + ", "
		}
	}
	selectStatement = fmt.Sprintf("SELECT %s%s ", aggregateListStatement, fieldListStatement)
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
		whereStatement = " WHERE " + strings.Join(query.Where, " ")
	}

	if query.OrderBy != nil {
		metadata := repository.DM.Metadatas[repository.Type]
		for key, value := range query.OrderBy {

			columnName, ok := metadata.FindColumnNameForField(key)
			if !ok {
				return "", nil, fmt.Errorf("No column found for field %s in OrderBy Query Part.", columnName)
			}
			if orderByStatement == "" {
				orderByStatement = fmt.Sprintf("%s %s", columnName, value)
			} else {
				orderByStatement = fmt.Sprintf("%s , %s %s ", orderByStatement, columnName, value)
			}
		}
		orderByStatement = " ORDER BY " + orderByStatement
	}
	if query.Limit != 0 {
		limitStatement = fmt.Sprintf(" LIMIT %d ", query.Limit)
	}
	if query.Offset != 0 {
		offsetStatement = fmt.Sprintf(" OFFSET %d ", query.Offset)
	}
	fromStatement = fmt.Sprintf(" FROM %s ", repository.TableName)
	q := []string{selectStatement, fromStatement, whereStatement, orderByStatement, limitStatement, offsetStatement, " ;"}
	return strings.Join(q, ""), values, nil
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
