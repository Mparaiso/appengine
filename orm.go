package orm

import (
	"fmt"
	"reflect"
)

type ORM struct {
	Connection  *Connection
	metadatas   map[reflect.Type]Metadata
	unityOfWork *UnityOfWork
}

func NewORM(connection *Connection) *ORM {
	return &ORM{connection, map[reflect.Type]Metadata{}, NewUnityOfWork()}
}

func (orm ORM) GetTypeMetadata(Type reflect.Type) Metadata {
	return orm.metadatas[Type]
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
func (orm ORM) Register(entities ...interface{}) error {
	for _, entity := range entities {
		if e, ok := entity.(MetadataProvider); ok {
			orm.metadatas[reflect.TypeOf(entity)] = e.DataMapperMetaData()
		} else {
			return fmt.Errorf("Cannot create metadata from Entity %#v .", entity)
		}
	}
	return nil
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

func (orm *ORM) GetConnection() *Connection {
	return orm.Connection
}

func (orm *ORM) UnityOfWork() *UnityOfWork {
	return orm.unityOfWork
}

func (orm *ORM) Persist(entities ...Entity) {

	for _, entity := range entities {

		if orm.resolveId(entity).(int64) == 0 {
			orm.UnityOfWork().Create(entity)
		} else {
			orm.UnityOfWork().Update(entity)
		}
	}
}

func (orm *ORM) Destroy(entities ...Entity) {
	for _, entity := range entities {
		orm.UnityOfWork().Delete(entity)
	}

}

func (orm *ORM) Flush() error {
	return orm.UnityOfWork().Flush(orm)
}

// resolveId gets and returns the value of the Primary Key column
// from the model
func (orm *ORM) resolveId(entity Entity) Any {
	Type := reflect.TypeOf(entity)
	value := reflect.Indirect(reflect.ValueOf(entity))
	return value.FieldByName(orm.GetTypeMetadata(Type).FindIdColumn().StructField).Interface()
}
