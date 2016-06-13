package orm

import "reflect"

type MetadataProvider interface {
	ProvideMetadata() Metadata
}

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
	ORM() *ORM
}

type QueryBuilder interface {
	GetType() QueryType
	BuildQuery(RepositoryInterface) (string, []interface{}, error)
}

type Logger interface {
	Log(arguments ...interface{})
}
