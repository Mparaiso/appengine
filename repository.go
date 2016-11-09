package orm

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
)

// Entity is a single entity that has
// metadatas
type Entity MetadataProvider

// Any is any type
type Any interface{}

// Collection is a collection of entities
type Collection interface{}

// Repository is a repository of entities of the same type
type Repository struct {
	Connection *Connection
	metadata   Metadata
	idField    string
	tableName  string
	aType      reflect.Type
	orm        *ORM
}

// TableName returns the table name for the repository
func (repository Repository) TableName() string {
	return repository.tableName
}

// Type return the type of entity managed by the repository
func (repository Repository) Type() reflect.Type {
	return repository.aType
}

// IDField returns the Id field in an entity managed by the repository
func (repository Repository) IDField() string {
	return repository.idField
}

// ORM returns the ORM associated with the repository
func (repository *Repository) ORM() *ORM {
	return repository.orm
}

func (repository *Repository) Metadata() Metadata {
	return repository.metadata
}

// NewRepository creates a new repository
func NewRepository(Type reflect.Type, orm *ORM) *Repository {
	metadata, ok := orm.metadatas[Type]
	if !ok {
		log.Fatalf("Datamapper cannot manage type %s", Type)
	}
	idField := metadata.FindIdColumn().Name
	if idField == "" {
		idField = metadata.FindIdColumn().Field
	}
	return &Repository{orm.Connection(), metadata, idField, metadata.Table.Name, Type, orm}
}

// All finds all
func (repository *Repository) All(collection Collection) error {
	return repository.FindBy(Query{}, collection)
}

// Find finds an entity by id
func (repository *Repository) Find(id Any, entity Entity) error {
	return repository.FindOneBy(Query{Where: []string{repository.idField, "=", "?"}, Params: []interface{}{id}, Limit: 1}, entity)
}

// FindOneBy finds one entity filtered by a query and resolve s
// relations defined in that entity
func (repository *Repository) FindOneBy(query Query, entity Entity) error {
	err := repository.doFindOneBy(query, entity)
	if err != nil {
		return err
	}
	// check for one to many relations
	metadata := repository.orm.metadatas[repository.aType]
	for _, relation := range metadata.Relations {
		if relation.Fetch == Eager {
			if relation.Type == OneToMany {
				err = repository.resolveOneToManySingle(relation, entity)
			}
			if relation.Type == OneToOne {
				err = repository.resolveOneToOneSingle(relation, entity)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FindBy find entities according to a query
func (repository *Repository) FindBy(query Query, collection Collection) error {
	err := repository.doFindBy(query, collection)
	if err != nil {
		return err
	}
	// check for one to many relations
	metadata := repository.orm.metadatas[repository.aType]
	for _, relation := range metadata.Relations {
		if relation.Fetch == Eager {
			if relation.Type == OneToMany {
				err = repository.resolveOneToMany(relation, collection)
			}
			if relation.Type == OneToOne {
				err = repository.resolveOneToOne(relation, collection)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Execute statement
func (repository *Repository) Execute(query Query) (sql.Result, error) {
	q, params, err := query.BuildQuery(repository)
	if err != nil {
		return nil, err
	}
	return repository.ORM().connection.Exec(q, params...)
}

// Count counts the number of rows
// that match the query
func (repository *Repository) Count(query Query) (int64, error) {
	query.Select = []string{""}
	query.Aggregates = []Aggregate{{Type: COUNT, StructField: "TOTAL", On: repository.idField}}
	queryString, values, err := query.BuildQuery(repository)
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
	result, err := repository.Connection.Exec(fmt.Sprintf("DELETE FROM %s;", repository.tableName))
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}
	return nil
}

// doFindBy finds entities without fetching related entities
func (repository *Repository) doFindBy(query Query, collection Collection) error {
	queryString, values, err := query.BuildQuery(repository)
	if err != nil {
		return err
	}
	err = repository.Connection.Select(collection, queryString, values...)

	return err
}

// doFindOneBy doesn't resolve relations between entities
func (repository *Repository) doFindOneBy(query Query, entity Entity) error {
	queryString, values, err := query.BuildQuery(repository)
	if err != nil {
		return err
	}
	return repository.Connection.Get(entity, queryString, values...)
}

// resolveId gets and returns the value of the Primary Key column
// from the model
func (repository *Repository) resolveID(entity Entity) Any {
	value := reflect.Indirect(reflect.ValueOf(entity))
	return value.FieldByName(repository.idField).Interface()
}

func (repository *Repository) resolveOneToMany(relation Relation, collection interface{}) error {

	if relation.Type == OneToMany {
		// The type name of the owned entity (string)
		ownedEntity := relation.TargetEntity
		// The repository that is used to fetch the owned
		ownedEntityRepository, err := repository.orm.GetRepositoryByEntityName(ownedEntity)
		if err != nil {
			return err
		}

		// Get all collection IDs
		ids := []interface{}{}
		// the reflect.Value of the collection, for reflection purposes
		collectionValue := reflect.Indirect(reflect.ValueOf(collection))
		// Fill with IDs to fetch
		collectionLength := collectionValue.Len()
		for i := 0; i < collectionLength; i++ {
			entityValue := reflect.Indirect(collectionValue.Index(i))
			ids = append(ids, entityValue.FieldByName(repository.idField).Interface())
		}
		// Build where query
		whereQuery := []string{relation.IndexedBy, "IN", "("}
		for range ids {
			whereQuery = append(whereQuery, "?", ",")
		}
		whereQuery[len(whereQuery)-1] = ")"

		// See http://stackoverflow.com/questions/25384640/why-golang-reflect-makeslice-returns-un-addressable-value
		sliceOfOwnedEntities := reflect.MakeSlice(reflect.SliceOf(ownedEntityRepository.aType), 0, 0)
		pointer := reflect.New(sliceOfOwnedEntities.Type())
		pointer.Elem().Set(sliceOfOwnedEntities)

		// Get All related target records in one request
		err = ownedEntityRepository.doFindBy(Query{Where: whereQuery, Params: ids, OrderBy: map[string]Order{relation.IndexedBy: ASC}}, pointer.Interface())
		if err != nil {
			return err
		}
		// assign target records to the returned collection entities
		sliceOfOwnedEntities = pointer.Elem()
		ownedEntitiesMap := map[interface{}][]reflect.Value{}
		sliceOfOwnedEntitiesLength := sliceOfOwnedEntities.Len()
		for i := 0; i < sliceOfOwnedEntitiesLength; i++ {
			ownedEntity := sliceOfOwnedEntities.Index(i)
			owningEntityIndex := reflect.Indirect(ownedEntity).FieldByName(relation.IndexedBy).Interface()
			if _, ok := ownedEntitiesMap[owningEntityIndex]; !ok {
				ownedEntitiesMap[owningEntityIndex] = []reflect.Value{}
			}
			ownedEntitiesMap[owningEntityIndex] = append(ownedEntitiesMap[owningEntityIndex], ownedEntity)
		}
		// add target records to each entity of the returned collection
		for i := 0; i < collectionLength; i++ {
			entityValue := reflect.Indirect(collectionValue.Index(i))
			// the slice value will contain the owned entities owned by the owning entity.
			sliceValue := reflect.MakeSlice(reflect.SliceOf(ownedEntityRepository.aType), 0, 0)
			sliceValue = reflect.Append(sliceValue, ownedEntitiesMap[entityValue.FieldByName(repository.idField).Interface()]...)
			entityValue.FieldByName(relation.Field).Set(sliceValue)
		}
		return nil
	}

	return fmt.Errorf("No relation to resolve for Relation '%v' for Type '%s' ", relation, repository.aType)
}

func (repository *Repository) resolveOneToManySingle(relation Relation, entity Entity) error {

	if relation.Type == OneToMany {
		targetRepository, err := repository.orm.GetRepositoryByEntityName(relation.TargetEntity)
		if err != nil {
			return err
		}
		// See http://stackoverflow.com/questions/25384640/why-golang-reflect-makeslice-returns-un-addressable-value
		slice := reflect.MakeSlice(reflect.SliceOf(targetRepository.aType), 0, 0)
		pointer := reflect.New(slice.Type())
		pointer.Elem().Set(slice)

		err = targetRepository.doFindBy(
			Query{Where: []string{relation.IndexedBy, "=", "?"},
				Params: []interface{}{repository.resolveID(entity)},
			}, pointer.Interface(),
		)
		if err != nil {
			return err
		}
		entityValue := reflect.Indirect(reflect.ValueOf(entity))
		entityValue.FieldByName(relation.Field).Set(pointer.Elem())
		return nil
	}

	return fmt.Errorf("No relation to resolve for Relation '%v' for Type '%s' ", relation, repository.aType)
}

func (repository *Repository) resolveOneToOne(relation Relation, collection interface{}) error {
	if relation.Type == OneToOne {
		collectionValue := reflect.Indirect(reflect.ValueOf(collection))
		if reflect.Slice != collectionValue.Kind() && reflect.Array != collectionValue.Kind() {
			return NotASliceError(fmt.Sprintf("Slice or Array expected, got %v", collection))
		}
		ids := []interface{}{}
		for i := 0; i < collectionValue.Len(); i++ {
			ids = append(ids, reflect.Indirect(collectionValue.Index(i)).FieldByName(repository.IDField()).Interface())
		}
		resultRepository, err := repository.ORM().GetRepositoryByEntityName(relation.TargetEntity)
		if err != nil {
			return err
		}
		sliceOfResultsType := reflect.SliceOf(resultRepository.Type())
		sliceOfResultValue := reflect.MakeSlice(sliceOfResultsType, 0, 0)
		pointer := reflect.New(sliceOfResultsType)
		pointer.Elem().Set(sliceOfResultValue)
		query := Query{Where: []string{relation.IndexedBy, "IN", "("}, Params: []interface{}{}}
		for _, id := range ids {
			query.Where = append(query.Where, "?", ",")
			query.Params = append(query.Params, id)
		}
		query.Where = append(query.Where[:len(query.Where)-1], ")")
		if err = resultRepository.FindBy(query, pointer.Interface()); err != nil {
			return err
		}
		sliceOfResultValue = reflect.Indirect(pointer)
		for i := 0; i < sliceOfResultValue.Len(); i++ {
			for j := 0; j < collectionValue.Len(); j++ {
				if target, entity := reflect.Indirect(sliceOfResultValue.Index(i)), reflect.Indirect(collectionValue.Index(j)); target.FieldByName(relation.IndexedBy).Interface() == entity.FieldByName(repository.IDField()).Interface() {
					entity.FieldByName(relation.Field).Set(sliceOfResultValue.Index(i))
					break
				}
			}
		}
	} else {
		return RelationNotHandled(fmt.Sprintf("The relation %v is not handled by ResolveOneToOne", relation))
	}
	return nil
}

func (repository *Repository) resolveOneToOneSingle(relation Relation, entity Entity) error {
	if relation.Type == OneToOne {
		entityValue := reflect.Indirect(reflect.ValueOf(entity))
		relatedEntityRepository, err := repository.ORM().GetRepositoryByEntityName(relation.TargetEntity)
		if err != nil {
			return err
		}
		slice := reflect.MakeSlice(reflect.SliceOf(relatedEntityRepository.Type()), 0, 0)
		pointer := reflect.New(slice.Type())
		pointer.Elem().Set(slice)
		// entityValue.FieldByName(relation.Field).Set(relatedValue)
		id := repository.resolveID(entity)
		err = relatedEntityRepository.FindBy(
			Query{Where: []string{relation.IndexedBy, "=", "?"}, Params: []interface{}{id}},
			pointer.Interface(),
		)
		if err != nil {
			return err
		}
		if pointer.Elem().Len() > 0 {
			entityValue.FieldByName(relation.Field).Set(pointer.Elem().Index(0))
		}
	}
	return nil
}

func (repository *Repository) LoadRelationSingle(entity Entity, fieldName string) error {
	if reflect.TypeOf(entity) != repository.Type() {
		return fmt.Errorf("The repository (%s) doesn't manage entities of type %v", repository.Type(), entity)
	}
	metadata := repository.ORM().GetEntityMetadata(entity)
	relation, resolved := metadata.ResolveRelationForFieldName(fieldName)
	if !resolved {
		return fmt.Errorf("Relation not found for field name '%s' in entity %s ", fieldName, entity)
	}
	switch relation.Type {
	case OneToOne:
		return repository.resolveOneToOneSingle(relation, entity)
	case OneToMany:
		return repository.resolveOneToManySingle(relation, entity)
	}
	return nil
}
