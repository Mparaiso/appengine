package orm

import (
	"fmt"
	"reflect"
)

type ORM struct {
	connection  *Connection
	metadatas   map[reflect.Type]Metadata
	unityOfWork *UnityOfWork
}

func NewORM(connection *Connection) *ORM {
	return &ORM{connection, map[reflect.Type]Metadata{}, NewUnityOfWork()}
}

func (orm ORM) GetTypeMetadata(Type reflect.Type) Metadata {
	return orm.metadatas[Type]
}

func (orm ORM) GetValueMetadata(value reflect.Value) Metadata {
	return orm.GetTypeMetadata(value.Type())
}

func (orm ORM) GetEntityMetadata(entity Entity) Metadata {
	return orm.GetTypeMetadata(reflect.TypeOf(entity))
}

func (orm ORM) HasTypeMetadata(Type reflect.Type) bool {
	_, ok := orm.metadatas[Type]
	return ok
}

func (orm ORM) HasEntityMetadata(entity Entity) bool {
	_, ok := orm.metadatas[reflect.TypeOf(entity)]
	return ok
}

// Register registers entities and make a repository available
// for each entity.
func (orm *ORM) Register(entities ...interface{}) error {
	for _, entity := range entities {
		if e, ok := entity.(MetadataProvider); ok {
			orm.metadatas[reflect.TypeOf(entity)] = e.ProvideMetadata()
		} else {
			return fmt.Errorf("Cannot create metadata from Entity %#v .", entity)
		}
	}
	return nil
}

func (orm *ORM) MustRegister(entities ...interface{}) {
	if err := orm.Register(entities...); err != nil {
		panic(err)
	}

}
func (orm *ORM) GetRepository(entity Entity) (*Repository, error) {
	Type := reflect.TypeOf(entity)
	if _, ok := orm.metadatas[Type]; ok {
		return NewRepository(Type, orm), nil
	}
	return nil, fmt.Errorf("Metadata not found for type %s .", Type)
}

// GetRepositoryByTableName gets a repository according to a table name.
func (orm *ORM) GetRepositoryByTableName(tableName string) (*Repository, error) {
	for Type, metadata := range orm.metadatas {
		if tableName == metadata.Table.Name {
			return NewRepository(Type, orm), nil
		}
	}
	return nil, fmt.Errorf("Repository not found for table '%s' .", tableName)
}

// GetRepositoryByEntityName find a repository by entity name
func (orm *ORM) GetRepositoryByEntityName(entityName string) (*Repository, error) {
	for Type, metadata := range orm.metadatas {
		if entityName == metadata.Entity {
			return NewRepository(Type, orm), nil
		}
	}
	return nil, fmt.Errorf("Repository not found for entity '%s' .", entityName)
}

func (orm *ORM) Connection() *Connection {
	return orm.connection
}

func (orm *ORM) UnityOfWork() *UnityOfWork {
	return orm.unityOfWork
}

// Persist inserts or updates an entity depending on
// the value of its primary key. If the primary key equals
// 0, a new entity will be inserted
func (orm *ORM) Persist(entities ...Entity) error {

	for _, entity := range entities {
		_, err := orm.GetRepository(entity)
		if err != nil {
			return err
		}
		if orm.resolveId(entity).(int64) == 0 {
			orm.UnityOfWork().Create(entity)
		} else {
			orm.UnityOfWork().Update(entity)
		}

		metadata := orm.GetEntityMetadata(entity)

		for _, relation := range metadata.Relations {

			if (relation.Cascade & Persist) != 0 {

				if relation.Type == OneToMany {
					entityValue := reflect.Indirect(reflect.ValueOf(entity))
					collectionValue := entityValue.FieldByName(relation.StructField)

					for i := 0; i < collectionValue.Len(); i++ {
						if e, ok := collectionValue.Index(i).Interface().(Entity); ok {
							orm.Persist(e)
						}
					}
				}
			}
		}
	}
	return nil
}

func (orm *ORM) MustPersist(entities ...Entity) *ORM {
	if err := orm.Persist(entities...); err != nil {
		panic(err)
	}
	return orm
}
func (orm *ORM) Remove(entities ...Entity) error {
	for _, entity := range entities {
		_, err := orm.GetRepository(entity)
		if err != nil {
			return err
		}
		metadata := orm.GetEntityMetadata(entity)
		for _, relation := range metadata.Relations {
			// Cascade remove.
			// The owned entities are removed BEFORE the owning entity
			// to prevent constraint violations like referencial integrity errors

			if (relation.Cascade & Remove) != 0 {
				if relation.Type == OneToMany {
					entityValue := reflect.Indirect(reflect.ValueOf(entity))
					collectionValue := entityValue.FieldByName(relation.StructField)
					for i := 0; i < collectionValue.Len(); i++ {
						if e, ok := collectionValue.Index(i).Interface().(Entity); ok {
							orm.Remove(e)
						}

					}
				}
			}
		}
		orm.UnityOfWork().Remove(entity)
	}
	return nil
}

func (orm *ORM) MustRemove(entities ...Entity) *ORM {
	if err := orm.Remove(entities...); err != nil {
		panic(err)
	}
	return orm
}
func (orm *ORM) Flush() error {
	return orm.UnityOfWork().Flush(orm)
}

func (orm *ORM) MustFlush() {
	if err := orm.UnityOfWork().Flush(orm); err != nil {
		panic(err)
	}
}

// resolveId gets and returns the value of the Primary Key column
// from the model
func (orm *ORM) resolveId(entity Entity) Any {
	Type := reflect.TypeOf(entity)
	value := reflect.Indirect(reflect.ValueOf(entity))
	return value.FieldByName(orm.GetTypeMetadata(Type).FindIdColumn().StructField).Interface()
}
