package datamapper

import (
	"fmt"
	"reflect"
	"strings"
)

type unityOfWork struct {
	deleted []*MetadataProvider
	updated []*MetadataProvider
	created []*MetadataProvider
}

func (u *unityOfWork) create(entities ...*MetadataProvider) {
	u.created = append(u.created, entities...)
}

func (u *unityOfWork) update(entities ...*MetadataProvider) {
	u.created = append(u.updated, entities...)
}

func (u *unityOfWork) delete(entities ...*MetadataProvider) {
	u.deleted = append(u.deleted, entities...)
}

func (u *unityOfWork) flush(dm *DataMapper) error {
	transaction, err := dm.Connection.db.Beginx()
	if err != nil {
		return err
	}
	for _, entity := range u.created {
		if e, ok := interface{}(entity).(BeforeCreateListener); ok == true {
			if err := e.BeforeCreate(); err != nil {
				transaction.Rollback()
				return err
			}
		}
		paths := []string{}
		values := []interface{}{}
		Type := reflect.TypeOf(entity)
		metadata := dm.Metadatas[Type]
		fieldMap := metadata.FieldMap(entity)
		idField := metadata.FindIdColumn().StructField
		tableName := metadata.Table.Name

		for key, value := range fieldMap {
			if strings.ToLower(key) != strings.ToLower(idField) {
				paths = append(paths, key)
				values = append(values, value.Interface())
			}
		}
		query := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s);",
			tableName,
			strings.Join(paths, ","),
			strings.Join(
				strings.Split(strings.Repeat("?", len(paths)), ""), ","))
		result, err := transaction.Exec(query, values...)
		if err != nil {
			transaction.Rollback()
			return err
		}
		lastInsertedId, err := result.LastInsertId()
		if err != nil {
			transaction.Rollback()
			return err
		}
		reflect.Indirect(reflect.ValueOf(entity)).FieldByName(metadata.FindIdColumn().StructField).SetInt(lastInsertedId)
	}
	return nil
}
