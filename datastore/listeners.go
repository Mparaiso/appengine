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
	"fmt"
	"time"
)

func BeforeEntityCreatedListener(e Event) error {
	switch event := e.(type) {
	case BeforeEntityCreatedEvent:
		if e, ok := event.Entity.(CreatedUpdatedEntity); ok {
			e.SetCreated(time.Now())
			e.SetUpdated(time.Now())
		}
		if e, ok := event.Entity.(VersionedEntity); ok {
			e.SetVersion(1)
		}
	}
	return nil
}

func BeforeEntityUpdatedListener(e Event) error {
	switch event := e.(type) {
	case BeforeEntityUpdatedEvent:
		if entity, ok := event.Old.(LockedEntity); ok {
			if entity.IsLocked() {
				return fmt.Errorf("Entity is locked and cannot be modified")
			}
		}
		if entity, ok := event.Old.(VersionedEntity); ok {
			if old, new := entity, event.New.(VersionedEntity); old.GetVersion() != new.GetVersion() {
				return fmt.Errorf("Versions do not match old : %d , new : %d", old.GetVersion(), new.GetVersion())
			} else {
				new.SetVersion(old.GetVersion() + 1)
			}
		}
		if entity, ok := event.New.(CreatedUpdatedEntity); ok {
			entity.SetUpdated(time.Now())
		}
		event.New.SetID(event.Old.GetID())
	}
	return nil
}

func BeforeEntityDeletedListener(e Event) error {
	switch event := e.(type) {
	case BeforeEntityDeletedEvent:
		if entity, ok := event.Entity.(LockedEntity); ok {
			if entity.IsLocked() {
				return fmt.Errorf("Entity is locked and cannot be modified")
			}
		}
	}
	return nil
}
