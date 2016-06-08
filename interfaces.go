package orm

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

type QueryBuilder interface {
	BuildQuery(*Repository) (string, []interface{}, error)
}

type Logger interface {
	Log(arguments ...interface{})
}
