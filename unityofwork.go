package orm

import (
	"fmt"
	"reflect"
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
	u.updated = append(u.updated, entities...)
}

func (u *UnityOfWork) Delete(entities ...MetadataProvider) {
	u.deleted = append(u.deleted, entities...)
}

func (u *UnityOfWork) Flush(dm *ORM) error {
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
		values := []interface{}{}
		Type := reflect.TypeOf(entity)
		metadata := dm.metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper", entity, Type)
		}
		repository,err:=dm.GetRepository(entity)
		if err!=nil{
			transaction.Rollback()
			return err
		}
		Set := metadata.BuildFieldValueMap(entity)
		query,values,err := Query{Type:INSERT,Set:Set}.BuildQuery(repository)
		if err!=nil{
			transaction.Rollback()
			return err
		}
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
		metadata := dm.metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper", entity, Type)
		}
		Set := metadata.BuildFieldValueMap(entity)
		IDField := metadata.FindIdColumn().StructField
		entityValue := reflect.Indirect(reflect.ValueOf(entity))
		ID := entityValue.FieldByName(IDField).Interface()
		repository, err := dm.GetRepository(entity)
		if err != nil {
			transaction.Rollback()
			return err
		}
		query, values, err := Query{
			Type:   UPDATE,
			Set:    Set,
			Where:  []string{IDField, "=", "?"},
			Params: []interface{}{ID},
		}.BuildQuery(repository)

		result, err := transaction.Exec(query, values...)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if _, err := result.RowsAffected(); err != nil {
			transaction.Rollback()
			return err
		}

	}
	// Delete entities
	for _, entity := range u.deleted {
		Type := reflect.TypeOf(entity)

		metadata := dm.metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper", entity, Type)
		}
		repository, err := dm.GetRepository(entity)
		if err != nil {
			transaction.Rollback()
			return err
		}
		idColumn := metadata.ResolveColumnNameFor(metadata.FindIdColumn())
		id := reflect.Indirect(reflect.ValueOf(entity)).FieldByName(metadata.FindIdColumn().StructField).Interface()
		query, values, err := Query{Type: DELETE,
			Where:  []string{idColumn, "=", "?"},
			Params: []interface{}{id},
		}.BuildQuery(repository)
		if err != nil {
			transaction.Rollback()
			return err
		}
		_, err = transaction.Exec(query, values...)
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
