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

	"github.com/Mparaiso/appengine/datastore"
	"google.golang.org/appengine"
)

type validationErrors interface {
	GetValidationErrors() map[string][]string
}

// CreatedMessage is returned when a new entity is created
type CreatedMessage struct {
	Status  int
	Message string
	ID      int64
}

// Resource is a reusable rest endpoint
type Resource struct {
	// Prototype is a value used to create other values
	Prototype datastore.Entity
	protoType reflect.Type
	// the datastore kind
	Kind          string
	Signal        datastore.Signal
	ResultPerPage int
	ErrorFunction func(writer http.ResponseWriter, Error error, status int)
}

// NewResource creates a new EndPoint
func NewResource(prototype datastore.Entity, Kind string) *Resource {
	return &Resource{Prototype: prototype, Kind: Kind}
}

// GetKind returns the endpoint's kind
func (e *Resource) GetKind() string {
	return e.Kind
}

// GetErrorFunction returns the function used to manage errors
func (e *Resource) GetErrorFunction() func(http.ResponseWriter, error, int) {
	if e.ErrorFunction == nil {
		e.ErrorFunction = func(w http.ResponseWriter, err error, status int) { http.Error(w, err.Error(), status) }
	}
	return e.ErrorFunction
}

// GetPrototype returns e.Prototype
func (e *Resource) GetPrototype() reflect.Type {
	if e.protoType == nil {
		e.protoType = reflect.Indirect(reflect.ValueOf(e.Prototype)).Type()
	}
	return e.protoType
}

// Index list resources
func (e Resource) Index(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	repository := datastore.NewDefaultRepositoryWithSignal(ctx, e.Kind, e.GetSignal())
	entities := reflect.New(reflect.SliceOf(e.GetPrototype())).Interface()
	err := repository.FindAll(entities)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(entities)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
	}
}

// GetSignal returns a signal dispatcher
func (e *Resource) GetSignal() datastore.Signal {
	if e.Signal == nil {
		e.Signal = datastore.NewDefaultSignal()
	}
	return e.Signal
}

// Get fetches a resource
func (e Resource) Get(w http.ResponseWriter, r *http.Request) {

	var id int64
	_, err := fmt.Sscanf(r.URL.Query().Get(":"+e.Kind), "%d", &id)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	entity := reflect.New(e.GetPrototype()).Interface().(datastore.Entity)
	ctx := appengine.NewContext(r)
	repository := datastore.NewDefaultRepositoryWithSignal(ctx, e.Kind, e.GetSignal())
	err = repository.FindByID(id, entity)
	if err == datastore.ErrNoSuchEntity {
		e.GetErrorFunction()(w, err, http.StatusNotFound)
		return
	} else if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(entity)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
	}
}

// Put updates a resource
func (e Resource) Put(w http.ResponseWriter, r *http.Request) {
	entity := reflect.New(e.GetPrototype()).Interface().(datastore.Entity)
	var id int64
	_, err := fmt.Sscanf(r.URL.Query().Get(":"+e.Kind), "%d", &id)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)

	repository := datastore.NewDefaultRepositoryWithSignal(ctx, e.Kind, e.GetSignal())
	err = repository.Update(entity)
	if validationErrors, ok := err.(validationErrors); ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors.GetValidationErrors())
		return
	} else if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Delete deletes a resource
func (e Resource) Delete(w http.ResponseWriter, r *http.Request) {

	entity := reflect.New(e.GetPrototype()).Interface().(datastore.Entity)
	var id int64
	_, err := fmt.Sscanf(r.URL.Query().Get(":"+e.Kind), "%d", &id)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)

	repository := datastore.NewDefaultRepositoryWithSignal(ctx, e.Kind, e.GetSignal())
	err = repository.Delete(entity)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// Post creates a resource
func (e Resource) Post(w http.ResponseWriter, r *http.Request) {
	entity := reflect.New(e.GetPrototype()).Interface()

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(entity)
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)

	repository := datastore.NewDefaultRepositoryWithSignal(ctx, e.Kind, e.GetSignal())

	err = repository.Create(entity.(datastore.Entity))
	if validationErrors, ok := err.(validationErrors); ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors.GetValidationErrors())
		return
	} else if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	type Created struct {
		Status  int
		Message string
		ID      int64
	}
	err = json.NewEncoder(w).Encode(CreatedMessage{Status: 201, Message: "Created", ID: entity.(datastore.Entity).GetID()})
	if err != nil {
		e.GetErrorFunction()(w, err, http.StatusInternalServerError)
	}
}
