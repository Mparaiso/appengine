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

package datastore

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

// Kind list app kinds
var Kind = struct{ Users, Migrations, Snippets, Categories, Roles, UserRoles string }{
	"Users", "Migrations", "Snippets", "Categories", "Roles", "UserRoles",
}

var _ Repository = new(DefaultRepository)

// DefaultRepository is the default implementation of Repository
type DefaultRepository struct {
	Context   context.Context
	Kind      string
	Signal    Signal
	ParentKey *datastore.Key
}

// NewDefaultRepositoryWithSignal allows to create a repository with an external signal
func NewDefaultRepositoryWithSignal(ctx context.Context, kind string, signal Signal) *DefaultRepository {
	defaultRepository := &DefaultRepository{Context: ctx, Kind: kind}
	defaultRepository.Signal = signal
	defaultRepository.Signal.Add(ListenerFunc(BeforeEntityCreatedListener))
	defaultRepository.Signal.Add(ListenerFunc(BeforeEntityUpdatedListener))
	return defaultRepository
}

// NewDefaultRepository creates a new DefaultRepository
func NewDefaultRepository(ctx context.Context, kind string, listeners ...Listener) *DefaultRepository {
	defaultRepository := &DefaultRepository{Context: ctx, Kind: kind}
	defaultRepository.Signal = NewDefaultSignal()
	defaultRepository.Signal.Add(ListenerFunc(BeforeEntityCreatedListener))
	defaultRepository.Signal.Add(ListenerFunc(BeforeEntityUpdatedListener))
	for _, listener := range listeners {
		defaultRepository.Signal.Add(listener)
	}
	return defaultRepository
}

type ContextValue int

const (
	ParentKey ContextValue = iota
)

var (
	ErrParentKeyNotFound = fmt.Errorf("ErrParentKeyNotFound")
	ErrNoSuchEntity      = datastore.ErrNoSuchEntity
	ErrNotAnEntity       = fmt.Errorf("This value doesn't implement Entity interface")
)

// SetParentKey sets the parent key
func (repository *DefaultRepository) SetParentKey(key *datastore.Key) {
	repository.ParentKey = key
}

// GetParentKey returns the parent key or nil if not set
func (repository DefaultRepository) GetParentKey() *datastore.Key {
	return repository.ParentKey
}

// Create an entity
func (repository DefaultRepository) Create(entity Entity) error {
	parentKey := repository.GetParentKey()
	low, _, err := datastore.AllocateIDs(repository.Context, repository.Kind, parentKey, 1)

	if err == nil {
		entity.SetID(low)
		err = repository.Dispatch(BeforeEntityCreatedEvent{Context: repository.Context, Entity: entity})
		if err != nil {
			return err
		}
		key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), parentKey)
		_, err = datastore.Put(repository.Context, key, entity)
		if err != nil {
			return err
		}
		err = repository.Dispatch(AfterEntityCreatedEvent{Context: repository.Context, Entity: entity})
	}
	return err
}

// Dispatch dispatches an event to the Signal if the Signal is not null
func (repository DefaultRepository) Dispatch(event Event) error {
	if repository.Signal != nil {
		return repository.Signal.Dispatch(event)
	}
	return nil
}

// CreateMulti persist multiple entities into the datastore
func (repository DefaultRepository) CreateMulti(entities ...Entity) error {
	parentKey := repository.GetParentKey()
	low, _, err := datastore.AllocateIDs(repository.Context, repository.Kind, parentKey, len(entities))
	keys := []*datastore.Key{}
	if err == nil {
		for i, entity := range entities {
			if e, ok := entity.(Entity); ok {
				e.SetID(low + int64(i))
				err = repository.Dispatch(BeforeEntityCreatedEvent{Context: repository.Context, Entity: e})
				if err != nil {
					return err
				}
				keys = append(keys, datastore.NewKey(repository.Context, repository.Kind, "", e.GetID(), parentKey))
			} else {
				return ErrNotAnEntity
			}
		}
		_, err = datastore.PutMulti(repository.Context, keys, entities)
		if err != nil {
			return err
		}
		for _, entity := range entities {
			err = repository.Dispatch(AfterEntityCreatedEvent{Context: repository.Context, Entity: entity})
			if err != nil {
				return err
			}
		}
	}
	return err
}

// Update an entity
func (repository DefaultRepository) Update(entity Entity) error {
	parentKey := repository.GetParentKey()

	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), parentKey)
	old := reflect.New(reflect.Indirect(reflect.ValueOf(entity)).Type()).Interface()
	err := datastore.Get(repository.Context, key, old)
	if err != nil {
		return err
	}
	err = repository.Dispatch(BeforeEntityUpdatedEvent{Context: repository.Context, Old: old.(Entity), New: entity})
	if err != nil {
		return err
	}
	_, err = datastore.Put(repository.Context, key, entity)
	if err != nil {
		return err
	}
	return repository.Dispatch(AfterEntityUpdatedEvent{Context: repository.Context, Old: old.(Entity), New: entity})

}

// Delete an entity
func (repository DefaultRepository) Delete(entity Entity) error {
	var err error
	parentKey := repository.GetParentKey()
	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), parentKey)
	err = repository.Dispatch(BeforeEntityDeletedEvent{Context: repository.Context, Entity: entity})
	if err != nil {
		return err
	}
	err = datastore.Delete(repository.Context, key)
	if err != nil {
		return err
	}
	return repository.Dispatch(AfterEntityDeletedEvent{Context: repository.Context, Entity: entity})
}

// FindByID gets an entity by id
func (repository DefaultRepository) FindByID(id int64, entity Entity) error {
	var err error
	parentKey := repository.GetParentKey()
	if err != nil {
		return err
	}
	key := datastore.NewKey(repository.Context, repository.Kind, "", id, parentKey)
	return datastore.Get(repository.Context, key, entity)
}

// FindAll returns all entities
func (repository DefaultRepository) FindAll(entities interface{}) error {
	parentKey := repository.GetParentKey()
	query := datastore.NewQuery(repository.Kind)
	if parentKey != nil {
		query = query.Ancestor(parentKey)
	}
	_, err := query.GetAll(repository.Context, entities)
	return err
}

type Query struct {
	Query  map[string]interface{}
	Order  []string
	Fields []string
	Limit  int
	Offset int
}

func (repository DefaultRepository) FindBy(
	query Query,
	result interface{}) error {
	parentKey := repository.GetParentKey()
	q := repository.createQuery(query)
	if parentKey != nil {
		q = q.Ancestor(parentKey)
	}
	_, err := q.GetAll(repository.Context, result)
	return err
}

// Count returns the object count given a query
func (repository DefaultRepository) Count(
	query Query) (int, error) {
	parentKey := repository.GetParentKey()
	q := repository.createQuery(query)
	if parentKey != nil {
		q = q.Ancestor(parentKey)
	}
	return q.Count(repository.Context)
}

func (repository DefaultRepository) createQuery(query Query) *datastore.Query {
	q := datastore.NewQuery(repository.Kind)
	for key, value := range query.Query {
		q = q.Filter(key, value)
	}
	for _, o := range query.Order {
		q = q.Order(o)
	}
	if len(query.Fields) > 0 {
		q = q.Project(query.Fields...)
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	return q.Offset(query.Offset)

}
