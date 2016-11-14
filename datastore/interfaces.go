//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package datastore

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// An Entity is a datastore entity
//
// GetID returns the ID
//
// SetID sets the ID
type Entity interface {
	GetID() int64
	SetID(int64)
}

// CreatedUpdatedSetter is capable to set created and updated timestamps on a struct
type CreatedUpdatedEntity interface {
	SetCreated(date time.Time)
	SetUpdated(date time.Time)
}
type VersionedEntity interface {
	SetVersion(int64)
	GetVersion() int64
}

type LockedEntity interface {
	IsLocked() bool
}

// ContextFactory creates a ContextProvider
type ContextFactory interface {
	Create(r *http.Request) context.Context
}

// ContextProvider provides a context
type ContextProvider interface {
	GetContext() context.Context
}
type RepositoryProvider interface {
	GetRepository() Repository
}

type SignalProvider interface {
	GetSignal() Signal
}
