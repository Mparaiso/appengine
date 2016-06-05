package datamapper

import (
	"encoding/json"
	"reflect"
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

type Cascade int8

const (
	_ Cascade = iota
	Persist
	Remove
	Merge
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
	ID          bool
	StructField string
	Name        string
}

// Relation Represents a relation between 2 entities
type Relation struct {
	Type         RelationType
	TargetEntity string
	MappedBy     string
	IndexBy      string
	Cascade
	Fetch
	StructField string
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
	var dmm Metadata
	err := json.Unmarshal([]byte(jsonString), &dmm)
	return dmm, err
}

func (dmm Metadata) FindIdColumn() Column {
	var column Column
	for _, value := range dmm.Columns {
		if value.ID {
			column = value
			break
		}
	}
	return column
}

func (dmm Metadata) FindColumnNameForField(fieldName string) (string, bool) {
	columnName := ""
	found := false
	for _, column := range dmm.Columns {
		if fieldName == column.StructField {
			if column.Name != "" {
				columnName = column.Name
			} else {
				columnName = column.StructField
			}
			found = true
			break
		}
	}
	return columnName, found
}

func (dmm Metadata) FieldMap(entity interface{}) (fieldMap map[string]reflect.Value) {
	value := reflect.Indirect(reflect.ValueOf(entity))
	fieldMap = map[string]reflect.Value{}
	for _, column := range dmm.Columns {
		name := column.Name
		if name == "" {
			name = column.StructField
		}
		fieldMap[name] = value.FieldByName(column.StructField)
	}
	return
}
