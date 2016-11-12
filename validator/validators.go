package validator

import (
	"fmt"

	"github.com/Mparaiso/appengine/datastore"
	"github.com/Mparaiso/go-tiger/validator"
)

// UniqueEntityValidator is valid when the validated entity is unique
type UniqueEntityValidator struct {
	datastore.Repository
}

func newUniqueEntityValidator(repository datastore.Repository) *UniqueEntityValidator {
	return &UniqueEntityValidator{repository}
}

// Validate does validate, values are used to find a potential dupblicate
func (provider UniqueEntityValidator) Validate(field string, values map[string]interface{}, errors validator.ValidationError) {
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
		errors.Append(field, "Should be unique")
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
func (provider EntityExistsValidator) Validate(field string, kind string, values map[string]interface{}, errors validator.ValidationError) {
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
