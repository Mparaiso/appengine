package tools

import (
	"fmt"
	"reflect"
)

// Scanner populates destination values
// or returns an error
type Scanner interface {
	Scan(destination ...interface{}) error
}

// MapRowToStruct  automatically maps a db row to a struct .
func MapRowToStruct(columns []string, scanner Scanner, Struct interface{}, ignoreMissingFields bool) error {
	structPointer := reflect.ValueOf(Struct)
	if structPointer.Kind() != reflect.Ptr {
		return fmt.Errorf("Pointer expected, got %#v", Struct)
	}
	structValue := reflect.Indirect(structPointer)
	zeroValue := reflect.Value{}
	arrayOfResults := []interface{}{}
	for _, column := range columns {
		field := structValue.FieldByName(column)
		if field == zeroValue {
			if ignoreMissingFields {
				pointer := reflect.New(reflect.TypeOf([]byte{}))
				pointer.Elem().Set(reflect.ValueOf([]byte{}))
				arrayOfResults = append(arrayOfResults, pointer.Interface())

			} else {
				return fmt.Errorf("No field found for column %s in struct %#v", column, Struct)

			}
		} else {
			if !field.CanSet() {
				return fmt.Errorf("Unexported field %s cannot be set in struct %#v", column, Struct)
			}
			arrayOfResults = append(arrayOfResults, field.Addr().Interface())
		}
	}
	err := scanner.Scan(arrayOfResults...)
	if err != nil {
		return err
	}
	return nil
}
