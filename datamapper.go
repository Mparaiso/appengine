package datamapper

import (
	"fmt"
	"reflect"
)

type DataMapper struct {
	Connection *Connection
	Metadatas  map[reflect.Type]Metadata
}

func NewDataMapper(connection *Connection) *DataMapper {
	return &DataMapper{connection, map[reflect.Type]Metadata{}}
}

// Register registers entities and make a repository available
// for each entity.
func (dm DataMapper) Register(entities ...interface{}) error {
	for _, entity := range entities {
		if e, ok := entity.(MetadataProvider); ok {
			dm.Metadatas[reflect.TypeOf(entity)] = e.DataMapperMetaData()
		} else {
			return fmt.Errorf("Cannot create metadata from Entity %#v .", entity)
		}
	}
	return nil
}

func (dm *DataMapper) GetRepository(entity Entity) (*Repository, error) {
	Type := reflect.TypeOf(entity)
	if _, ok := dm.Metadatas[Type]; ok {
		return NewRepository(Type, dm), nil
	}
	return nil, fmt.Errorf("Metadata not found for type %s .", Type)
}

// GetRepositoryByTableName gets a repository according to a table name.
func (dm *DataMapper) GetRepositoryByTableName(tableName string) (*Repository, error) {
	for Type, metadata := range dm.Metadatas {
		if tableName == metadata.Table.Name {
			return NewRepository(Type, dm), nil
		}
	}
	return nil, fmt.Errorf("Repository not found for table '%s' .", tableName)
}

// GetRepositoryByEntityName find a repository by entity name
func (dm *DataMapper) GetRepositoryByEntityName(entityName string) (*Repository, error) {
	for Type, metadata := range dm.Metadatas {
		if entityName == metadata.Entity {
			return NewRepository(Type, dm), nil
		}
	}
	return nil, fmt.Errorf("Repository not found for entity '%s' .", entityName)
}

func (dm *DataMapper) GetConnection() *Connection {
	return dm.Connection
}
