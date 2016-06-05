package datamapper

import (
	"fmt"
	"log"
	"reflect"
)

// UserRepository is a repository of users
type Repository struct {
	Connection *Connection
	IDField    string
	TableName  string
	Type       reflect.Type
	DM         *DataMapper
}

// Entity is a single entity that has
// metadatas
type Entity MetadataProvider

// Any is any type
type Any interface{}

// Collection is a collection of entities
type Collection interface{}

func NewRepository(Type reflect.Type, datamapper *DataMapper) *Repository {
	metadata, ok := datamapper.Metadatas[Type]
	if !ok {
		log.Fatalf("Datamapper cannot manage type %s", Type)
	}
	idField := metadata.FindIdColumn().Name
	if idField == "" {
		idField = metadata.FindIdColumn().StructField
	}
	return &Repository{datamapper.GetConnection(), idField, metadata.Table.Name, Type, datamapper}
}

// All finds all
func (repository *Repository) All(collection Collection) error {
	return repository.FindBy(Query{}, collection)
}

// Find finds an entity by id
func (repository *Repository) Find(id Any, entity Entity) error {
	return repository.FindOneBy(Query{Where: []string{repository.IDField, "=", "?"}, Params: []interface{}{id}}, entity)
}

// FindOneBy find one record filtered by query
func (repository *Repository) FindOneBy(query QueryBuilder, entity Entity) error {
	queryString, values, err := query.AcceptRepository(repository)
	err = repository.Connection.Get(entity, queryString, values...)
	if err != nil {
		return err
	}
	// check for one to many relations
	metadata := repository.DM.Metadatas[repository.Type]
	for _, relation := range metadata.Relations {
		if relation.Fetch == Eager {
			if relation.Type == OneToMany {
				targetRepository, err := repository.DM.GetRepositoryByEntityName(relation.TargetEntity)
				if err != nil {
					return err
				}
				// See http://stackoverflow.com/questions/25384640/why-golang-reflect-makeslice-returns-un-addressable-value
				slice := reflect.MakeSlice(reflect.SliceOf(targetRepository.Type), 0, 0)
				pointer := reflect.New(slice.Type())
				pointer.Elem().Set(slice)

				err = targetRepository.FindBy(
					Query{Where: []string{relation.IndexBy, "=", "?"},
						Params: []interface{}{repository.resolveId(entity)},
					}, pointer.Interface(),
				)

				if err != nil {
					return err
				}
				StructFieldValue := reflect.Indirect(reflect.ValueOf(entity)).FieldByName(relation.StructField)

				StructFieldValue.Set(slice)
			}
		}
	}
	return nil
}

func (repository *Repository) doFindOneBy(query QueryBuilder, entity Collection) error {
	return nil
}

// FindBy find users by fields
func (repository *Repository) FindBy(query QueryBuilder, collection Collection) error {
	queryString, values, err := query.AcceptRepository(repository)
	if err != nil {
		return err
	}

	err = repository.Connection.Select(collection, queryString, values...)
	if err != nil {
		return err
	}
	// check for one to many relations
	metadata := repository.DM.Metadatas[repository.Type]
	for _, relation := range metadata.Relations {
		if relation.Fetch == Eager {
			if relation.Type == OneToMany {
				targetEntity := relation.TargetEntity
				targetRepository, err := repository.DM.GetRepositoryByEntityName(targetEntity)
				if err != nil {
					return err
				}

				// Get all collection IDs
				ids := []interface{}{}
				collectionValue := reflect.Indirect(reflect.ValueOf(collection))
				collectionLength := collectionValue.Len()
				for i := 0; i < collectionLength; i++ {
					entityValue := reflect.Indirect(collectionValue.Index(i))
					ids = append(ids, entityValue.FieldByName(repository.IDField).Interface())
				}
				// Build where query
				whereQuery := []string{relation.IndexBy, "IN", "("}
				for range ids {
					whereQuery = append(whereQuery, "?", ",")
				}
				whereQuery[len(whereQuery)-1] = ")"

				// See http://stackoverflow.com/questions/25384640/why-golang-reflect-makeslice-returns-un-addressable-value
				slice := reflect.MakeSlice(reflect.SliceOf(targetRepository.Type), 0, 0)
				pointer := reflect.New(slice.Type())
				pointer.Elem().Set(slice)

				// Get All related target records in one request
				err = targetRepository.FindBy(Query{Where: whereQuery, Params: ids, OrderBy: map[string]string{relation.IndexBy: "ASC"}}, pointer.Interface())
				if err != nil {
					return err
				}
				// assign target records to the returned collection entities
				targetRecordMap := map[interface{}][]reflect.Value{}
				targetSliceLength := slice.Len()
				for i := 0; i < targetSliceLength; i++ {
					targetRecord := slice.Index(i)
					index := targetRecord.FieldByName(relation.IndexBy).Interface()
					if _, ok := targetRecordMap[index]; !ok {
						targetRecordMap[index] = []reflect.Value{}
					}
					targetRecordMap[index] = append(targetRecordMap[index], targetRecord)
				}
				// add target records to each entity of the returned collection
				for i := 0; i < collectionLength; i++ {
					entityValue := reflect.Indirect(collectionValue.Index(i))
					sliceValue := reflect.MakeSlice(reflect.SliceOf(targetRepository.Type), 0, 0)
					sliceValue = reflect.Append(sliceValue, targetRecordMap[entityValue.FieldByName(repository.IDField).Interface()]...)
					entityValue.FieldByName(relation.StructField).Set(sliceValue)
				}
			}
		}
	}
	return nil
}

// Count counts the number of rows
// that match the query
func (repository *Repository) Count(query Query) (int64, error) {
	query.Select = []string{""}
	query.Aggregates = []Aggregate{{Type: COUNT, StructField: "TOTAL", On: repository.IDField}}
	queryString, values, err := query.AcceptRepository(repository)
	if err != nil {
		return 0, err
	}
	var result int64
	// @TODO modify Query so it can support aggregation and only return specific parts
	err = repository.Connection.Get(&result, fmt.Sprintf(queryString), values...)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// DeleteAll deletes all models
func (repository *Repository) DeleteAll() error {
	result, err := repository.Connection.Exec(fmt.Sprintf("DELETE FROM %s;", repository.TableName))
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}
	return nil
}

// Save persists an entity.
func (repository *Repository) Save(entities ...Entity) error {
	repository.DM.Persist(entities...)
	return repository.DM.Flush()
}

// UpdateAttribute update selected attributes
func (repository *Repository) UpdateAttribute(entity Entity, fields map[string]interface{}) error {
	values := []interface{}{}
	setStatement := ""
	userType := reflect.TypeOf(entity)
	for key, value := range fields {
		if _, ok := userType.Elem().FieldByName(key); !ok {
			return fmt.Errorf("type %s doesn't have a field named %s ", userType.Name(), key)
		}
		values = append(values, value)
		if setStatement == "" {
			setStatement = fmt.Sprintf(" %s = ?", key)
		} else {
			setStatement = fmt.Sprintf(" %s, %s = ?", setStatement, key)
		}
	}
	id := repository.resolveId(entity)
	result, err := repository.Connection.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE ID = ?;", repository.TableName, setStatement), append(values, id)...)
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}
	err = repository.Find(id, entity)
	if err != nil {
		return err
	}
	return nil
}

// Destroy deletes an entity
func (repository *Repository) Destroy(entity Entity) error {
	id := repository.resolveId(entity)
	result, err := repository.Connection.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE %s.%s=?", repository.TableName, repository.IDField),
		id,
	)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err != nil {
		return err
	} else if rows <= 0 {
		return fmt.Errorf("User with ID %d could not be found and desftroyed.", id)
	}
	return nil
}

// resolveId gets and returns the value of the Primary Key column
// from the model
func (repository *Repository) resolveId(entity Entity) Any {
	value := reflect.Indirect(reflect.ValueOf(entity))
	return value.FieldByName(repository.IDField).Interface()
}
