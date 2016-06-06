package orm

type NotImplementedError string

func (n NotImplementedError) Error() string {
	return string(n)
}
