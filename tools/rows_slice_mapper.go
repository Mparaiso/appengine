package tools

import (
	"fmt"
	"reflect"
)

type RowsScanner interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(destination ...interface{}) error
}

func MapRowsToSliceOfStruct(scanner RowsScanner, sliceOfStructs interface{}, ignoreMissingField bool) error {
	///return connection.db.Select(records, query, parameters...)
	recordsPointerValue := reflect.ValueOf(sliceOfStructs)
	if recordsPointerValue.Kind() != reflect.Ptr {
		return fmt.Errorf("Expect pointer, got %#v", sliceOfStructs)
	}
	recordsValue := recordsPointerValue.Elem()
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("The underlying type is not a slice,pointer to slice expected for %#v ", recordsValue)
	}
	defer scanner.Close()
	columns, err := scanner.Columns()
	if err != nil {
		return err
	}
	// get the underlying type of a slice
	// @see http://stackoverflow.com/questions/24366895/golang-reflect-slice-underlying-type
	for scanner.Next() {
		//
		var t reflect.Type
		if recordsValue.Type().Elem().Kind() == reflect.Ptr {
			// the sliceOfStructs type is like []*T
			t = recordsValue.Type().Elem().Elem()
		} else {
			// the sliceOfStructs type is like []T
			t = recordsValue.Type().Elem()
		}
		pointerOfElement := reflect.New(t)
		err = MapRowToStruct(columns, scanner, pointerOfElement.Interface(), ignoreMissingField)
		if err != nil {
			return err
		}
		recordsValue = reflect.Append(recordsValue, pointerOfElement)
	}
	recordsPointerValue.Elem().Set(recordsValue)
	return scanner.Err()
}
