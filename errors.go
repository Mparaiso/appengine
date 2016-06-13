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

type NotASliceError string

func (n NotASliceError) Error() string {
	if t := string(n); t == "" {
		return "Slice or Array expected"
	} else {
		return t
	}
}

type RelationNotHandled string

func (n RelationNotHandled) Error() string {
	return string(n)
}
