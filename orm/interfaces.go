package orm

import "reflect"

type MetadataProvider interface {
	ProvideMetadata() Metadata
}

type Entity MetadataProvider

type Any interface{}

type BeforeCreateListener interface {
	BeforeCreate() error
}

type AfterCreateListener interface {
	AfterCreate() error
}
type BeforeUpdateListener interface {
	BeforeUpdate() error
}
type AfterUpdateListener interface {
	AfterUpdate() error
}

type BeforeRemoveListener interface {
	BeforeRemove() error
}

type AfterRemoveListener interface {
	AfterRemove() error
}

type RepositoryInterface interface {
	TableName() string
	IDField() string
	Type() reflect.Type
	ORM() ORMInterface
	Metadata() Metadata
}

type ORMInterface interface {
	GetMetadataByEntityName(string) (Metadata, bool)
	GetTypeMetadata(reflect.Type) Metadata
	Metadatas() map[reflect.Type]Metadata
}

type QueryBuilderInterface interface {
	BuildQuery(RepositoryInterface) (string, []interface{}, error)
}

type Logger interface {
	Log(arguments ...interface{})
}
