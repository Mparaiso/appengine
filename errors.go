package orm

type NotImplementedError string

func (n NotImplementedError) Error() string {
	return string(n)
}

type EntityNotRegisiteredInORMError string

func (n EntityNotRegisiteredInORMError) Error() string {
	return string(n)
}

type NotAPointerError string

func (n NotAPointerError) Error() string {
	return string(n)
}
