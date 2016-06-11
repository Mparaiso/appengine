package orm

import "reflect"

type MetadataProvider interface {
	ProvideMetadata() Metadata
}

type BeforeCreateListener interface {
	BeforeCreate() error
}
type BeforeUpdateListener interface {
	BeforeUpdate() error
}

type BeforeDestroyListener interface {
	BeforeDestroy() error
}

type RepositoryInterface interface {
	TableName() string
	IDField() string
	Type() reflect.Type
	ORM() *ORM
}

type QueryBuilder interface {
	GetType() QueryType
	BuildQuery(RepositoryInterface) (string, []interface{}, error)
}

type Logger interface {
	Log(arguments ...interface{})
}
