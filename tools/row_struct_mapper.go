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
			//pointer := reflect.New(field.Type())
			//pointer.Elem().Set(field)
			//arrayOfResults = append(arrayOfResults, pointer.Interface())
			//Can replaced by the following using reflect.Value.Addr()
			arrayOfResults = append(arrayOfResults, field.Addr().Interface())
		}
	}
	err := scanner.Scan(arrayOfResults...)
	if err != nil {
		return err
	}
	// Not Needed anymore since we are pointing directly to the Struct fields
	/*
		valueOfResults := reflect.ValueOf(arrayOfResults)
		for index, column := range columns {
			field := structValue.FieldByName(column)
			if field == zeroValue {
				return fmt.Errorf("No field found for column %s in struct %#v", column, Struct)
			}
			field.Set(valueOfResults.Index(index).Elem().Elem())
		}
	*/
	//structPointer.Elem().Set(structValue)
	return nil
}
