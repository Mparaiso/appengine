package datamapper

type MetadataProvider interface {
	DataMapperMetaData() Metadata
}

type BeforeCreateListener interface {
	BeforeCreate() error
}

type BeforeSaveListener interface {
	BeforeSave() error
}

type BeforeUpdateListener interface {
	BeforeUpdate() error
}

type BeforeDestroyListener interface {
	BeforeDestroy() error
}

type QueryBuilder interface {
	AcceptRepository(*Repository) (string, []interface{}, error)
}

type Logger interface {
	Log(arguments ...interface{})
}
