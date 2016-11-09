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
	"golang.org/x/net/context"
)

type BeforeEntityDeletedEvent struct {
	Context context.Context
	Entity
}
type AfterEntityDeletedEvent struct {
	Context context.Context
	Entity
}
type BeforeEntityUpdatedEvent struct {
	Context context.Context
	Old     Entity
	New     Entity
}
type AfterEntityUpdatedEvent struct {
	Context context.Context
	Old     Entity
	New     Entity
}

type BeforeEntityCreatedEvent struct {
	Context context.Context
	Entity
}

type AfterEntityCreatedEvent struct {
	Context context.Context
	Entity
}
