package orm

import (
	"fmt"
	"regexp"
	"strings"

	. "github.com/mparaiso/go-orm/query"
)

// Query implements QueryBuilder.
// It can be used to filter entities when they are fetched
// from the database.
type QueryBuilder struct {
	Type       Type
	Select     []string
	From       []string
	Where      []string
	Params     []interface{}
	OrderBy    map[string]Order
	Join       []Join
	Limit      int64
	Offset     int64
	Aggregates []Aggregate
	Set        map[string]interface{}
}

func (query QueryBuilder) GetType() Type {
	return query.Type
}

// BuildQuery builds the querystring with placeholders , returns it with the parameters
// to replace the placeholders and an error.
func (builder QueryBuilder) BuildQuery(orm ORMInterface) (string, []interface{}, error) {
	query := NewBuilder()
	switch builder.Type {
	case SELECT:
		// FROM
		for _, entity := range builder.From {
			metadata, ok := orm.GetMetadataByEntityName(strings.Trim(entity, " "))
			if !ok {
				return "", []interface{}{}, fmt.Errorf("The entity type %v is not registered.", entity)
			}
			query.From = append(query.From, metadata.TableName())
		}
		// SELECT
		err := builder.buildSelect(orm, query)
		if err != nil {
			return "", []interface{}{}, err
		}
		// WHERE

		err = builder.buildWhere(orm, query)
		if err != nil {
			return "", []interface{}{}, err
		}
		// LIMIT
		query.Limit = builder.Limit
		// OFFSET
		query.Offset = builder.Offset
		// ORDER BY
		for k, v := range builder.OrderBy {
			query.OrderBy[k] = v
		}
		query.OrderBy = builder.OrderBy

	default:
		return "", []interface{}{}, fmt.Errorf("The query type %v is not supported.", query.Type)

	}
	/*
		case SELECT:
			fromStatement := query.buildFromStatement(orm)
			joinStatement := ""
			for _, join := range query.Join {
				relation, resolved := orm.Metadata().ResolveRelationForTargetEntity(join.TargetEntity)
				if !resolved {
					return "", []interface{}{}, fmt.Errorf("Unresloved relation in join %v, with entity %s", join, join.TargetEntity)
				}
				targetMetadata, found := orm.GetMetadataByEntityName(join.TargetEntity)
				if !found {
					return "", []interface{}{}, fmt.Errorf("No metadata found for '%s'", join.TargetEntity)
				}
				var joinTableName, joinColumn, entityTableName, entityColumn string
				switch relation.Type {
				case ManyToOne:
					joinTableName = targetMetadata.TableName()
					joinColumn = targetMetadata.ResolveColumnNameFor(targetMetadata.FindIdColumn())
					entityTableName = orm.TableName()
					entityColumn = orm.Metadata().ResolveColumnNameByFieldName(relation.InversedBy)
				default:
					return "", []interface{}{}, fmt.Errorf("Relation of '%s' type '%s' for entity '%' is not handled in join statement", relation.Field, relation.Type, orm.Metadata().Entity)
				}

				joinStatement += "JOIN " + joinTableName + " ON " + joinTableName +
					"." + joinColumn + " = " + entityTableName + "." + entityColumn + " "
			}
			whereStatement, values, err := query.buildWhereStatement(orm)
			if err != nil {
				return "", values, err
			}
			selectStatement, err := query.buildSelectStatement(orm)
			if err != nil {
				return "", values, err
			}
			orderByStatement, err := query.buildOrderByStatment(orm)
			if err != nil {
				return "", values, err
			}
			limitStatement := query.buildLimitStatement(orm)
			offsetStatement := query.buildOffsetStatement(orm)
			q := []string{selectStatement, fromStatement, joinStatement, whereStatement, orderByStatement, limitStatement, offsetStatement, " ;"}
			return strings.Join(q, ""), values, nil

		case UPDATE:
			updateStatement, values := query.buildUpdateStatement(orm)
			whereStatement, v, err := query.buildWhereStatement(orm)
			if err != nil {
				return "", values, err
			}
			values = append(values, v...)
			q := strings.Join([]string{updateStatement, whereStatement, ";"}, "")
			return q, values, nil

		case DELETE:
			deleteStatement := fmt.Sprintf("DELETE FROM %s ", orm.TableName())
			whereStatement, values, err := query.buildWhereStatement(orm)
			if err != nil {
				return "", values, err
			}
			limitStatement := query.buildLimitStatement(orm)
			query := deleteStatement + whereStatement + limitStatement + ";"
			return query, values, nil

		case INSERT:
			return query.buildInsertStatment(orm)

		default:
			return "", []interface{}{}, fmt.Errorf("The query type %v is not supported.", query.Type)
		}
	*/
	return query.BuildQuery()
}

/**
func (query QueryBuilder) buildInsertStatment(orm ORMInterface) (string, []interface{}, error) {
	values := []interface{}{}
	columns := []string{}
	idColumn := orm.IDField()
	tableName := orm.TableName()
	metadata := orm.ORM().GetTypeMetadata(orm.Type())
	for key, value := range query.Set {
		if key != idColumn {
			column := metadata.ResolveColumnNameByFieldName(key)
			if column == "" {
				return "", values, fmt.Errorf("No Column found for Field %s in Set for INSERT query", key)
			}
			columns = append(columns, column)
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

func (query QueryBuilder) buildUpdateStatement(orm ORMInterface) (string, []interface{}) {
	setStatement := ""
	values := []interface{}{}
	metadata := orm.ORM().GetTypeMetadata(orm.Type())
	fieldMap := query.Set
	for key, value := range fieldMap {
		columnName := metadata.ResolveColumnNameByFieldName(key)
		setStatement = fmt.Sprintf("%s %s = ? ,", setStatement, columnName)
		values = append(values, value)
	}
	setStatement = strings.TrimLeft(strings.TrimSuffix(setStatement, " ,"), " ")
	updateStatement := fmt.Sprintf("UPDATE %s SET %s", metadata.TableName(), setStatement)
	return updateStatement, values
}

func (query QueryBuilder) buildLimitStatement(orm ORMInterface) string {
	if query.Limit != 0 {
		return fmt.Sprintf(" LIMIT %d ", query.Limit)
	}
	return ""
}

func (query QueryBuilder) buildOffsetStatement(orm ORMInterface) string {
	if query.Offset != 0 {
		return fmt.Sprintf(" OFFSET %d ", query.Offset)
	}
	return ""
}

func (query QueryBuilder) buildOrderByStatment(orm ORMInterface) (string, error) {
	orderByStatement := ""
	if query.OrderBy != nil {
		metadata := orm.ORM().GetTypeMetadata(orm.Type())
		for key, value := range query.OrderBy {

			columnName := metadata.ResolveColumnNameByFieldName(key)
			if columnName == "" {
				return "", fmt.Errorf("No column found for field %s in OrderBy Query Part.", columnName)
			}
			if orderByStatement == "" {
				orderByStatement = fmt.Sprintf("%s %s", columnName, value)
			} else {
				orderByStatement = fmt.Sprintf("%s , %s %s ", orderByStatement, columnName, value)
			}
		}
		return " ORDER BY " + orderByStatement, nil
	}
	return "", nil
}
*/
func (query QueryBuilder) buildSelect(orm ORMInterface, builder *Builder) error {
	if len(query.From) == 0 {
		return fmt.Errorf("query.From is empty, no Entity to select from")
	}
	// if no Select or Select == ["*"] , get all fields from entity referenced by query.From[0]
	if len(query.Select) == 0 || (len(query.Select) == 1 && query.Select[0] == strings.Trim("*", " ")) {
		metadata, ok := orm.GetMetadataByEntityName(strings.Trim(query.From[0], " "))
		if !ok {
			return fmt.Errorf("Metadata not found for Entity", query.From[0])
		}
		tableName := metadata.TableName()
		for _, column := range metadata.Columns {
			builder.Select = append(builder.Select, tableName+"."+column.Name+" AS "+column.Field)
		}
		// len(Select)>0 , translate the fields into their respective column names
	} else {
		pattern := regexp.MustCompile(`(?i)^(\w+\.)?(\w+)(\s+AS\s+)?(\w+)?`)
		for _, selected := range query.Select {
			selected = strings.Trim(selected, " ")
			if found := pattern.FindStringSubmatch(selected); len(found) == 5 {
				entityName := strings.TrimRight(found[1], ".")
				fieldName := found[2]
				alias := found[4]
				if entityName == "" {
					entityName = query.From[0]
				} else if indexOfString(query.From, entityName) < 0 {
					return fmt.Errorf("Error in Select : '%s' : Entity not found in QueryBuilder.From '%v' ", selected, query.From)
				}
				metadata, ok := orm.GetMetadataByEntityName(entityName)
				if !ok {
					return fmt.Errorf("Error in Select: '%s' : Metadata not found for Entity '%s' in QueryBuilder.From '%v' ", selected, entityName, query.From)
				}
				column := metadata.ResolveColumnNameByFieldName(fieldName)
				if column == "" {
					return fmt.Errorf("Error in Select: '%s' : column not found for field '%s' in Entity '%s' in QueryBuilder.From '%v' ", selected, fieldName, entityName, query.From)
				}
				if alias == "" {
					alias = fieldName
				}
				builder.Select = append(builder.Select, metadata.TableName()+"."+column+" AS "+alias)
			}
		}
	}
	return nil

}

func (query QueryBuilder) buildWhere(orm ORMInterface, builder *Builder) error {
	if len(query.Where) != 0 {
		builder.Params = append(builder.Params, query.Params...)
		for _, token := range query.Where {
			switch token {
			case "=", "<", "<=", ">", ">=", "<>", "!=", "IN", "NOT IN", "NOT LIKE", "LIKE":
			default:
				builder.Where = append(builder.Where,token)
			}
		}
	}
	return nil
}

/*
func (query QueryBuilder) buildWhereStatement(orm ORMInterface) (string, []interface{}, error) {
	// values to be added as query parameters
	values := []interface{}{}
	metadata := orm.Metadata()
	if len(query.Where) > 0 {
		if len(query.Params) > 0 {
			values = append(values, query.Params...)
		}
		paramNumber := 0
		for index, token := range query.Where {
			switch token {
			case "?":
				paramNumber = paramNumber + 1
			case "=", "<", "<=", ">", ">=", "<>", "!=", "IN", "NOT IN", "NOT LIKE", "LIKE":
				var columnName, tableName string
				if index == 0 {
					return "", nil, fmt.Errorf("Unexpected token %s at index %d in Query.Where", token, index)
				}
				fieldName := query.Where[index-1]
				// if the field name like "Entity.Field"
				if fieldParts := strings.Split(fieldName, "."); len(fieldParts) > 1 {
					entityName, fieldName := fieldParts[0], fieldParts[1]
					metadata, found := orm.ORM().GetMetadataByEntityName(entityName)
					if !found {
						return "", nil, fmt.Errorf("Entity '%s' not found for field '%s' in Where Query Part", entityName, fieldName)
					}
					tableName = metadata.TableName()
					columnName = metadata.ResolveColumnNameByFieldName(fieldName)
					if columnName == "" {
						return "", nil, fmt.Errorf("No column found for field '%s' in Entity '%s' in Where Query Part.", fieldName, entityName)
					}
				} else {
					// the fieldname is like "field"
					columnName = metadata.ResolveColumnNameByFieldName(fieldName)
					tableName = metadata.TableName()
					if columnName == "" {
						return "", nil, fmt.Errorf("No column found for field '%s' in entity '%s' in Where Query Part.", fieldName, tableName)
					}
				}
				query.Where[index-1] = tableName + "." + columnName
			}

		}
		if paramNumber != len(values) {
			return "", nil, fmt.Errorf("Not enough ? placeholders for params %v in %s ", values, query.Where)
		}
		return "WHERE " + strings.Join(query.Where, " "), values, nil
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
		if (len(fieldNameSelector) > 0 && indexOfString(fieldNameSelector, column.Field) >= 0) || len(fieldNameSelector) == 0 {
			fields = fields + metadata.TableName() + "." + metadata.ResolveColumnNameByFieldName(column.Field) + " AS " + column.Field + ","
		}
	}
	// we remove the last comma in the string

	return string(fields[:len(fields)-1]), nil
}
**/
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
