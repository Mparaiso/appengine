package datamapper

import (
	"fmt"
	"reflect"
	"strings"
)

type UnityOfWork struct {
	deleted []MetadataProvider
	updated []MetadataProvider
	created []MetadataProvider
}

func NewUnityOfWork() *UnityOfWork {
	return &UnityOfWork{[]MetadataProvider{}, []MetadataProvider{}, []MetadataProvider{}}
}

func (u *UnityOfWork) Create(entities ...MetadataProvider) {
	u.created = append(u.created, entities...)
}

func (u *UnityOfWork) Update(entities ...MetadataProvider) {
	u.created = append(u.updated, entities...)
}

func (u *UnityOfWork) Delete(entities ...MetadataProvider) {
	u.deleted = append(u.deleted, entities...)
}

func (u *UnityOfWork) Flush(dm *DataMapper) error {
	// See http://stackoverflow.com/questions/24318389/golang-elem-vs-indirect-in-the-reflect-package

	transaction, err := dm.Connection.BeginTransaction()
	if err != nil {
		return err
	}
	// Create entities
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
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper.", entity, Type)
		}
		fieldMap := metadata.FieldMap(entity)
		idColumn := metadata.ResolveColumnNameFor(metadata.FindIdColumn())
		tableName := metadata.Table.Name

		for key, value := range fieldMap {
			if strings.ToLower(key) != strings.ToLower(idColumn) {
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
	// Update entities
	for _, entity := range u.updated {
		if e, ok := interface{}(entity).(BeforeUpdateListener); ok == true {
			if err := e.BeforeUpdate(); err != nil {
				transaction.Rollback()
				return err
			}
		}
		values := []interface{}{}
		Type := reflect.TypeOf(entity)
		metadata := dm.Metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper.", entity, Type)
		}
		fieldMap := metadata.FieldMap(entity)
		idColumn := metadata.ResolveColumnNameFor(metadata.FindIdColumn())
		tableName := metadata.Table.Name
		setStatement := "SET"
		var id interface{}
		for key, value := range fieldMap {
			if strings.ToLower(key) != strings.ToLower(idColumn) {
				setStatement = fmt.Sprintf("%s %s = ? ,", setStatement, key)
				values = append(values, value.Interface())
			} else {
				id = value.Interface()
			}
		}

		query := fmt.Sprintf("UPDATE %s WHERE %s = ? ;",
			tableName,
			setStatement,
			idColumn)
		result, err := transaction.Exec(query, append(values, id)...)
		if _, err := result.RowsAffected(); err != nil {
			transaction.Rollback()
			return err
		}
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	// Delete entities
	for _, entity := range u.deleted {
		Type := reflect.TypeOf(entity)
		metadata := dm.Metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper.", entity, Type)
		}
		tableName := metadata.Table.Name
		idColumn := metadata.ResolveColumnNameFor(metadata.FindIdColumn())
		_, err := transaction.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ? LIMIT 1 ;", tableName, idColumn),
			reflect.Indirect(reflect.ValueOf(entity)).FieldByName(metadata.FindIdColumn().StructField).Interface())
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	// Commit transaction
	err = transaction.Commit()
	if err != nil {
		return err
	}
	// Reset unity of work
	u.created = []MetadataProvider{}
	u.updated = []MetadataProvider{}
	u.deleted = []MetadataProvider{}
	return nil
}
