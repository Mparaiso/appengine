package validator

import (
	"fmt"

	"github.com/Mparaiso/appengine/datastore"
)

// ValidationError allows to collect
// multiple errors from different validators.
// This interface was ripped out of github.com/Mparaiso/go-tiger/validator
// and made independant
type ValidationError interface {
	// Append a new error to the errors map
	Append(key, value string)
}

// UniqueEntityValidator is valid when the validated entity is unique
type UniqueEntityValidator struct {
	datastore.Repository
}

// NewUniqueEntityValidator creates a new UniqueEntityValidator
func NewUniqueEntityValidator(repository datastore.Repository) *UniqueEntityValidator {
	return &UniqueEntityValidator{repository}
}

// Validate does validate, values are used to find a potential dupblicate
// example :
//
//		validation.Validate("Username",map[string]interface{}{"Username":user.Username},errors)
//
// will add an error to errors if another entity in the datastore has the same Username field value
func (provider UniqueEntityValidator) Validate(field string, values map[string]interface{}, errors ValidationError) {
	query := map[string]interface{}{}
	for key, value := range values {
		query[key+"="] = value
	}
	count, err := provider.Repository.Count(datastore.Query{Query: query, Limit: 1})
	if err != nil {
		errors.Append(field, err.Error())
		return

	}
	if count != 0 {
		errors.Append(field, "is already taken.")
	}

}

// EntityExistsValidator validates the fact
// that the targeted entity exists in the database
type EntityExistsValidator struct {
	datastore.Repository
}

// NewEntityExistsValidator returns an new EntityExistsValidator
func NewEntityExistsValidator(repository datastore.Repository) *EntityExistsValidator {
	return &EntityExistsValidator{repository}
}

// Validate does validate
func (provider EntityExistsValidator) Validate(field string, kind string, values map[string]interface{}, errors ValidationError) {
	query := map[string]interface{}{}
	for key, value := range values {
		query[key+"="] = value
	}
	count, err := provider.Repository.Count(datastore.Query{Query: query, Limit: 1})
	if err != nil {
		errors.Append(field, err.Error())
		return
	}
	if count == 0 {
		errors.Append(field, fmt.Sprintf("%s with fields matching %v does not exist", kind, values))
	}
}
