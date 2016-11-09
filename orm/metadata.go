package orm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// RelationType is an enum
type RelationType int8

const (
	_ RelationType = iota
	ManyToOne
	ManyToMany
	OneToOne
	OneToMany
)

// String returns the name of the relation type as a string
func (relation RelationType) String() string {
	switch relation {
	case ManyToOne:
		return "ManyToOne"
	case ManyToMany:
		return "ManyToMany"
	case OneToOne:
		return "OneToOne"
	case OneToMany:
		return "OneToMany"
	default:
		return ""
	}
}

// Cascade represents how related datas are handled
// when then entity is persisted or removed
type Cascade byte

const (
	// Persist automatically persists related entities when an entity is saves
	Persist Cascade = 0x01
	// Remove automatically removes related entities when an entity is removed
	Remove Cascade = 0x02
	// Merge   Cascade = 0x04
)

type Fetch int8

const (
	_ Fetch = iota
	Lazy
	ExtraLazy
	Eager
)

// Table is a  table metadata
type Table struct {
	Name string
}

// Column is a column metadata
type Column struct {
	ID    bool
	Field string
	Name  string
}

type JoinColumn struct {
	// The referenced id of the owning entity
	ReferencedField string
	// the foreign key of the owned entity
	Field string
}

// Relation Represents a relation between 2 entities
type Relation struct {
	// The Type of relation
	Type RelationType
	// The entity which is the target of the relation
	TargetEntity string
	MappedBy     string
	IndexedBy    string
	// The InversedBy attribute designates the field in the entity that is the inverse side of the relationship.
	InversedBy string
	Cascade
	// Whether the related entities are loaded
	// automatically or not.
	Fetch
	// The field where to load the related entities
	// if needed.
	Field string
	// For a OneToOne relationship
	JoinColumn JoinColumn
}

// DataMapperMetadata represent metadatas for a DB Table
type Metadata struct {
	Entity    string
	Table     Table
	Columns   []Column
	Relations []Relation
}

// MetadataFrom creates a datamapper metadata from a json string
// or returns an error
func MetadataFrom(jsonString string) (Metadata, error) {
	var meta Metadata
	err := json.Unmarshal([]byte(jsonString), &meta)
	return meta, err
}

// TableName returns the name of the table associated with the entity
func (meta Metadata) TableName() string {
	return strings.ToLower(meta.Table.Name)
}
func (meta Metadata) FindIdColumn() Column {
	var column Column
	for _, value := range meta.Columns {
		if value.ID {
			column = value
			break
		}
	}
	return column
}

func (meta Metadata) ResolveRelationForFieldName(fieldName string) (Relation, bool) {
	for _, relation := range meta.Relations {
		if relation.Field == fieldName {
			return relation, true
		}
	}
	return Relation{}, false
}

func (meta Metadata) ResolveColumnNameByFieldName(fieldName string) string {
	columnName := ""
	for _, column := range meta.Columns {
		if fieldName == column.Field {
			if column.Name != "" {
				columnName = column.Name
			} else {
				columnName = column.Field
			}
			break
		}
	}
	return strings.ToLower(columnName)
}

func (meta Metadata) BuildFieldValueMap(entity interface{}) (map[string]interface{}, error) {
	Set := map[string]interface{}{}
	entityValue := reflect.Indirect(reflect.ValueOf(entity))
	for _, column := range meta.Columns {
		if fieldValue := reflect.Indirect(entityValue.FieldByName(column.Field)); fieldValue != (reflect.Value{}) {
			Set[column.Field] = fieldValue.Interface()
		} else {
			fmt.Errorf("Field '%s' not found for column '%s' in entity '%s'", column.Field, meta.ResolveColumnNameFor(column), meta.Entity)
		}

	}
	return Set, nil
}

func (meta Metadata) FieldMap(entity interface{}) (fieldMap map[string]reflect.Value) {
	value := reflect.Indirect(reflect.ValueOf(entity))
	fieldMap = map[string]reflect.Value{}
	for _, column := range meta.Columns {
		fieldMap[meta.ResolveColumnNameFor(column)] = value.FieldByName(column.Field)
	}
	return
}

func (meta Metadata) ResolveColumnNameFor(column Column) (name string) {
	if column.Name == "" {
		name = column.Field
	} else {
		name = column.Name
	}
	return strings.ToLower(name)
}

func (meta Metadata) ResolveRelationForTargetEntity(entityName string) (Relation, bool) {
	if strings.Trim(entityName, " \n\t\r") != "" {
		for _, relation := range meta.Relations {
			if entityName == relation.TargetEntity {
				return relation, true
			}
		}
	}
	return Relation{}, false
}
