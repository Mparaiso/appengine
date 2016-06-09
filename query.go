package orm

import (
	"fmt"
	"strings"
)

// Order is an order
type Order string

const (
	// ASC in an ORDER BY statement
	ASC Order = "ASC"
	// DESC in an ORDER BY statement
	DESC Order = "DESC"
)

type QueryType int8

const (
	SELECT QueryType = iota
	INSERT
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

// Aggregate describes an aggregate operation like COUNT, SUM, MIN, MAX, AVG
type Aggregate struct {
	Type aggregateType
	On   string
	// StructField is the struct field that should be populated with the result of the aggregate
	StructField string
	// Separator used by GROUP_CONCAT
	Separator string
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
	Set        map[string]interface{}
}

// BuildQuery builds the querystring with placeholders , returns it with the parameters
// to replace the placeholders and an error.
func (query Query) BuildQuery(repository *Repository) (string, []interface{}, error) {

	switch query.Type {

	case SELECT:
		fromStatement := query.buildFromStatement(repository)
		whereStatement, values, err := query.buildWhereStatement(repository)
		if err != nil {
			return "", values, err
		}
		selectStatement, err := query.buildSelectStatement(repository)
		if err != nil {
			return "", values, err
		}
		orderByStatement, err := query.buildOrderByStatment(repository)
		if err != nil {
			return "", values, err
		}
		limitStatement := query.buildLimitStatement(repository)
		offsetStatement := query.buildOffsetStatement(repository)
		q := []string{selectStatement, fromStatement, whereStatement, orderByStatement, limitStatement, offsetStatement, " ;"}
		return strings.Join(q, ""), values, nil

	case UPDATE:
		updateStatement, values := query.buildUpdateStatement(repository)
		whereStatement, v, err := query.buildWhereStatement(repository)
		if err != nil {
			return "", values, err
		}
		values = append(values, v...)
		q := strings.Join([]string{updateStatement, whereStatement, ";"}, "")
		return q, values, nil

	case DELETE:
		deleteStatement := fmt.Sprintf("DELETE FROM %s ", repository.TableName)
		whereStatement, values, err := query.buildWhereStatement(repository)
		if err != nil {
			return "", values, err
		}
		limitStatement := query.buildLimitStatement(repository)
		query := deleteStatement + whereStatement + limitStatement + ";"
		return query, values, nil

	case INSERT:
		return query.buildInsertStatment(repository)

	default:
		return "", []interface{}{}, fmt.Errorf("The query type %v is not supported.", query.Type)
	}

}

func (query Query) buildInsertStatment(repository *Repository) (string, []interface{}, error) {
	values := []interface{}{}
	columns := []string{}
	idColumn := repository.IDField
	tableName := repository.TableName
	metadata := repository.ORM.GetTypeMetadata(repository.Type)
	for key, value := range query.Set {
		if key != idColumn {
			column := metadata.ResolveColumnNameByFieldName(key)
			if column == "" {
				return "", values, fmt.Errorf("No Column found for Field %s in Set for INSERT query", key)
			}
			columns = append(columns, strings.ToLower(column))
			values = append(values, value)
		}
	}
	q := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s);",
		tableName,
		strings.Join(columns, ","),
		strings.Join(
			strings.Split(strings.Repeat("?", len(columns)), ""), ","))
	return q, values, nil
}

func (query Query) buildUpdateStatement(repository *Repository) (string, []interface{}) {
	setStatement := ""
	values := []interface{}{}
	metadata := repository.ORM.GetTypeMetadata(repository.Type)
	fieldMap := query.Set
	for key, value := range fieldMap {
		columnName := strings.ToLower(metadata.ResolveColumnNameByFieldName(key))
		setStatement = fmt.Sprintf("%s %s = ? ,", setStatement, columnName)
		values = append(values, value)
	}
	setStatement = strings.TrimLeft(strings.TrimSuffix(setStatement, " ,"), " ")
	updateStatement := fmt.Sprintf("UPDATE %s SET %s", metadata.Table.Name, setStatement)
	return updateStatement, values
}

func (query Query) buildLimitStatement(repository *Repository) string {
	if query.Limit != 0 {
		return fmt.Sprintf(" LIMIT %d ", query.Limit)
	}
	return ""
}

func (query Query) buildOffsetStatement(repository *Repository) string {
	if query.Offset != 0 {
		return fmt.Sprintf(" OFFSET %d ", query.Offset)
	}
	return ""
}

func (query Query) buildOrderByStatment(repository *Repository) (string, error) {
	orderByStatement := ""
	if query.OrderBy != nil {
		metadata := repository.ORM.metadatas[repository.Type]
		for key, value := range query.OrderBy {

			columnName := metadata.ResolveColumnNameByFieldName(key)
			if columnName == "" {
				return "", fmt.Errorf("No column found for field %s in OrderBy Query Part.", columnName)
			}
			if orderByStatement == "" {
				orderByStatement = fmt.Sprintf("%s %s", strings.ToLower(columnName), value)
			} else {
				orderByStatement = fmt.Sprintf("%s , %s %s ", orderByStatement, strings.ToLower(columnName), value)
			}
		}
		return " ORDER BY " + orderByStatement, nil
	}
	return "", nil
}

func (query Query) buildSelectStatement(repository *Repository) (string, error) {
	metadata := repository.ORM.GetTypeMetadata(repository.Type)
	buildFromStatement := ""
	// Get columns to be returned ( like " table.column1 AS structField1 , table.column2 AS structField2 " )
	fieldListStatement, err := buildSelectFieldListFromColumnMetadata(metadata, query.Select...)
	if err != nil {
		return "", err
	}
	// Create aggregation statements ( like "COUNT(TABLE.COLUMN1) AS ALIAS" )
	if len(query.Aggregates) > 0 {
		for _, aggregate := range query.Aggregates {
			columnName := metadata.ResolveColumnNameByFieldName(aggregate.On)
			if columnName == "" {
				return "", fmt.Errorf("StructField '%s' Not Found on aggregate %v .", aggregate.On, aggregate)
			}
			buildFromStatement = buildFromStatement + " " + string(aggregate.Type) + "(" + repository.TableName + "." + columnName + ") AS " + aggregate.StructField
		}
		if fieldListStatement != "" {
			buildFromStatement = buildFromStatement + ", "
		}
	}
	return fmt.Sprintf("SELECT %s%s ", buildFromStatement, fieldListStatement), nil
}

func (query Query) buildFromStatement(repository *Repository) string {
	return fmt.Sprintf(" FROM %s ", repository.TableName)
}

func (query Query) buildWhereStatement(repository *Repository) (string, []interface{}, error) {
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
				columnName := metadata.ResolveColumnNameByFieldName(fieldName)
				if columnName == "" {
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
