package orm

import (
	"fmt"
	"reflect"
)

// UnityOfWork is a unity of work.
// Please use NewUnityOfWork to create
// a one, and not the struct directly
type UnityOfWork struct {
	deleted []MetadataProvider
	updated []MetadataProvider
	created []MetadataProvider
}

// NewUnityOfWork returns a new UnityOfWork
func NewUnityOfWork() *UnityOfWork {
	return &UnityOfWork{[]MetadataProvider{}, []MetadataProvider{}, []MetadataProvider{}}
}

// Create add an entity to the create list
func (u *UnityOfWork) Create(entities ...MetadataProvider) {
	// This will guarantee that no entity
	// exists in the unity of work twice
	for _, entity := range entities {
		u.Detach(entity)
	}
	u.created = append(u.created, entities...)
}

// Update adds an entity to the update list
func (u *UnityOfWork) Update(entities ...MetadataProvider) {
	// This will guarantee that no entity
	// exists in the unity of work twice
	for _, entity := range entities {
		u.Detach(entity)
	}
	u.updated = append(u.updated, entities...)
}

// Remove adds an entity to the delete list
func (u *UnityOfWork) Remove(entities ...MetadataProvider) {
	// This will guarantee that no entity
	// exists in the unity of work twice
	for _, entity := range entities {
		u.Detach(entity)
	}
	u.deleted = append(u.deleted, entities...)
}

// Detach remove entities from the unity of work
func (u *UnityOfWork) Detach(entity MetadataProvider) {
	for i, candidate := range u.deleted {
		if candidate == entity {
			u.deleted = append(u.deleted[:i], u.deleted[i+1:]...)
			return
		}
	}
	for i, candidate := range u.updated {
		if candidate == entity {
			u.deleted = append(u.updated[:i], u.updated[i+1:]...)
			return
		}
	}
	for i, candidate := range u.created {
		if candidate == entity {
			u.deleted = append(u.created[:i], u.created[i+1:]...)
			return
		}
	}
}

// Flush execute all operations in a single transaction
// then reset the unity of work.
// Returns an error if needed
func (u *UnityOfWork) Flush(orm *ORM) error {
	// See http://stackoverflow.com/questions/24318389/golang-elem-vs-indirect-in-the-reflect-package

	transaction, err := orm.connection.Begin()
	if err != nil {
		return err
	}
	// Create entities
	for _, entity := range u.created {
		repository, err := orm.GetRepository(entity)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if e, ok := interface{}(entity).(BeforeCreateListener); ok == true {
			if err := e.BeforeCreate(); err != nil {
				transaction.Rollback()
				return err
			}
		}
		values := []interface{}{}
		Type := reflect.TypeOf(entity)
		metadata := orm.metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper", entity, Type)
		}

		Set := metadata.BuildFieldValueMap(entity)
		query, values, err := Query{Type: INSERT, Set: Set}.BuildQuery(repository)
		if err != nil {
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
		repository, err := orm.GetRepository(entity)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if e, ok := interface{}(entity).(BeforeUpdateListener); ok == true {
			if err := e.BeforeUpdate(); err != nil {
				transaction.Rollback()
				return err
			}
		}
		values := []interface{}{}
		Type := reflect.TypeOf(entity)
		metadata := orm.metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper", entity, Type)
		}
		Set := metadata.BuildFieldValueMap(entity)
		IDField := metadata.FindIdColumn().StructField
		entityValue := reflect.Indirect(reflect.ValueOf(entity))
		ID := entityValue.FieldByName(IDField).Interface()
		delete(Set, IDField)
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
		repository, err := orm.GetRepository(entity)
		if err != nil {
			transaction.Rollback()
			return err
		}
		metadata := orm.metadatas[Type]
		if &metadata == nil {
			transaction.RollBack()
			return fmt.Errorf("entity '%#v' of type '%#v' is not managed by the datamapper", entity, Type)
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
