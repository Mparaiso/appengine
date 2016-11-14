//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"golang.org/x/net/context"

	"github.com/Mparaiso/appengine/datastore"
	"google.golang.org/appengine"
)

// Entity is a datastore entity
type Entity interface {
	GetID() int64
	SetID(int64)
}

// Validator valides an entity or return an error if the entity is invalid.
type Validator func(cxt context.Context, r *http.Request, entity Entity) error

// CreatedMessage is returned when a new entity is created
type CreatedMessage struct {
	Status  int
	Message string
	ID      int64
}

// Resource is a reusable rest endpoint
type Resource struct {
	// Prototype is a value used to create other values
	Prototype       Entity
	CreatePrototype Entity
	UpdatePrototype Entity
	protoType       reflect.Type
	// the datastore kind
	Kind          string
	Signal        datastore.Signal
	ResultPerPage int
	ErrorFunction func(writer http.ResponseWriter, Error error, status int)
	validator     Validator
}

// GetCreatePrototype returns resource.CreatePrototype
func (resource Resource) GetCreatePrototype() Entity {
	return resource.CreatePrototype
}

// SetCreatePrototype sets resource.CreatePrototype
func (resource *Resource) SetCreatePrototype(CreatePrototype Entity) {
	resource.CreatePrototype = CreatePrototype
}

// GetUpdatePrototype returns resource.UpdatePrototype
func (resource Resource) GetUpdatePrototype() Entity {
	return resource.UpdatePrototype
}

// SetUpdatePrototype sets resource.UpdatePrototype
func (resource *Resource) SetUpdatePrototype(UpdatePrototype Entity) {
	resource.UpdatePrototype = UpdatePrototype
}

// NewResource creates a new EndPoint
func NewResource(prototype Entity, Kind string) *Resource {
	return &Resource{Prototype: prototype, Kind: Kind}
}

// SetValidator sets a validator that can be used to validate an entity before it is persisted
// the validator MUST return the entity AND nil if no error
func (resource *Resource) SetValidator(validator Validator) {
	resource.validator = validator
}

// GetValidator returns resource.Validator
func (resource Resource) GetValidator() Validator {
	return resource.validator
}

// Validate validates an entity before datastore persistance
func (resource Resource) Validate(ctx context.Context, r *http.Request, entity Entity) error {
	if validator := resource.GetValidator(); validator == nil {
		return nil
	} else {
		return validator(ctx, r, entity)
	}
}

// GetKind returns the endpoint's kind
func (resource *Resource) GetKind() string {
	return resource.Kind
}

// GetErrorFunction returns the function used to manage errors
func (resource *Resource) GetErrorFunction() func(http.ResponseWriter, error, int) {
	if resource.ErrorFunction == nil {
		resource.ErrorFunction = func(w http.ResponseWriter, err error, status int) { http.Error(w, err.Error(), status) }
	}
	return resource.ErrorFunction
}

// GetPrototype returns r.Prototype
func (r *Resource) GetPrototype() reflect.Type {
	if r.protoType == nil {
		r.protoType = reflect.Indirect(reflect.ValueOf(r.Prototype)).Type()
	}
	return r.protoType
}

// Index list resources
func (resource Resource) Index(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	repository := datastore.NewDefaultRepositoryWithSignal(ctx, resource.Kind, resource.GetSignal())
	entities := reflect.New(reflect.SliceOf(resource.GetPrototype())).Interface()
	err := repository.FindAll(entities)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(entities)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
	}
}

// GetSignal returns a signal dispatcher
func (r *Resource) GetSignal() datastore.Signal {
	if r.Signal == nil {
		r.Signal = datastore.NewDefaultSignal()
	}
	return r.Signal
}

// Get fetches a resource
func (resource Resource) Get(w http.ResponseWriter, r *http.Request) {

	var id int64
	_, err := fmt.Sscanf(r.URL.Query().Get(":"+resource.Kind), "%d", &id)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	entity := reflect.New(resource.GetPrototype()).Interface().(Entity)
	ctx := appengine.NewContext(r)
	repository := datastore.NewDefaultRepositoryWithSignal(ctx, resource.Kind, resource.GetSignal())
	err = repository.FindByID(id, entity)
	if err == datastore.ErrNoSuchEntity {
		resource.GetErrorFunction()(w, err, http.StatusNotFound)
		return
	} else if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(entity)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
	}
}

// Put updates a resource
func (resource Resource) Put(w http.ResponseWriter, r *http.Request) {
	var id int64
	_, err := fmt.Sscanf(r.URL.Query().Get(":"+resource.Kind), "%d", &id)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	entity := reflect.New(resource.GetPrototype()).Interface().(Entity)
	err = json.NewDecoder(r.Body).Decode(entity)
	entity.SetID(id)
	ctx := appengine.NewContext(r)
	if err = resource.Validate(ctx, r, entity); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}
	repository := datastore.NewDefaultRepositoryWithSignal(ctx, resource.Kind, resource.GetSignal())
	err = repository.Update(entity)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Delete deletes a resource
func (resource Resource) Delete(w http.ResponseWriter, r *http.Request) {

	var id int64
	_, err := fmt.Sscanf(r.URL.Query().Get(":"+resource.Kind), "%d", &id)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	entity := reflect.New(resource.GetPrototype()).Interface().(Entity)
	entity.SetID(id)
	ctx := appengine.NewContext(r)

	repository := datastore.NewDefaultRepositoryWithSignal(ctx, resource.Kind, resource.GetSignal())
	err = repository.Delete(entity)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// Post creates a resource
func (resource Resource) Post(w http.ResponseWriter, r *http.Request) {
	entity := reflect.New(resource.GetPrototype()).Interface().(Entity)

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(entity)
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)

	repository := datastore.NewDefaultRepositoryWithSignal(ctx, resource.Kind, resource.GetSignal())
	if err = resource.Validate(ctx, r, entity); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}
	err = repository.Create(entity.(Entity))
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	type Created struct {
		Status  int
		Message string
		ID      int64
	}
	err = json.NewEncoder(w).Encode(CreatedMessage{Status: 201, Message: "Created", ID: entity.(Entity).GetID()})
	if err != nil {
		resource.GetErrorFunction()(w, err, http.StatusInternalServerError)
	}
}
