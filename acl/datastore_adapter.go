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

package acl

import (
	"encoding/gob"
	"time"

	"github.com/Mparaiso/appengine/datastore"
	tiger_acl "github.com/Mparaiso/go-tiger/acl"
	"golang.org/x/net/context"
)

func init() {
	gob.Register(tiger_acl.Allow)
}

// Rule is an ACL rule
type Rule struct {
	ID            int64
	Created       time.Time
	Updated       time.Time
	Type          tiger_acl.Type
	RoleID        string
	ResourceID    string
	AllPrivileges bool
	Assertion     tiger_acl.Assertion `datastore:"-"`
	Privilege     string
}

// GetID returns a int64
func (rule Rule) GetID() int64 {
	return rule.ID
}

// SetID sets *Rule.rule
func (rule *Rule) SetID(ID int64) {
	rule.ID = ID
}

// GetCreated returns a time.Time
func (rule Rule) GetCreated() time.Time {
	return rule.Created
}

// SetCreated sets *Rule.rule
func (rule *Rule) SetCreated(Created time.Time) {
	rule.Created = Created

}

// GetUpdated returns a time.Time
func (rule Rule) GetUpdated() time.Time {
	return rule.Updated
}

// SetUpdated sets *Rule.rule
func (rule *Rule) SetUpdated(Updated time.Time) {
	rule.Updated = Updated

}

type ResourceNode struct {
	ID               int64
	ResourceID       string
	ParentResourceID string
}

/*
 * Getters and setters for struct type ResourceNode
 */

// GetID returns a int64
func (resourceNode ResourceNode) GetID() int64 {
	return resourceNode.ID
}

// SetID sets *ResourceNode.resourceNode
func (resourceNode *ResourceNode) SetID(ID int64) {
	resourceNode.ID = ID

}

// GetResourceID returns a string
func (resourceNode ResourceNode) GetResourceID() string {
	return resourceNode.ResourceID
}

// SetResourceID sets *ResourceNode.resourceNode and returns
func (resourceNode *ResourceNode) SetResourceID(ResourceID string) {
	resourceNode.ResourceID = ResourceID

}

// GetParentResourceID returns a string
func (resourceNode ResourceNode) GetParentResourceID() string {
	return resourceNode.ParentResourceID
}

// SetParentResourceID sets *ResourceNode.resourceNode and returns
func (resourceNode *ResourceNode) SetParentResourceID(ParentResourceID string) {
	resourceNode.ParentResourceID = ParentResourceID

}

type RoleNode struct {
	ID           int64
	ParentRoleID string
	RoleID       string
}

// GetID returns a int64
func (roleNode RoleNode) GetID() int64 {
	return roleNode.ID
}

// SetID sets *RoleNode.roleNode and returns *RoleNode
func (roleNode *RoleNode) SetID(ID int64) {
	roleNode.ID = ID

}

// GetParentRoleID returns a string
func (roleNode RoleNode) GetParentRoleID() string {
	return roleNode.ParentRoleID
}

// SetParentRoleID sets *RoleNode.roleNode and returns *RoleNode
func (roleNode *RoleNode) SetParentRoleID(ParentRoleID string) {
	roleNode.ParentRoleID = ParentRoleID

}

// GetRoleID returns a string
func (roleNode RoleNode) GetRoleID() string {
	return roleNode.RoleID
}

// SetRoleID sets *RoleNode.roleNode and returns *RoleNode
func (roleNode *RoleNode) SetRoleID(RoleID string) {
	roleNode.RoleID = RoleID

}

type DatastoreAdapter struct {
	ctx context.Context
	ACL *tiger_acl.ACL
	RoleNodesKind,
	ResourceNodesKind,
	RulesKind string
}

const (
	ResourceNodesKind = "acl_resource_nodes"
	RoleNodesKind     = "acl_role_nodes"
	RulesKind         = "acl_rules"
)

func NewDatastoreAdapter(ctx context.Context) *DatastoreAdapter {
	adapter := &DatastoreAdapter{ctx: ctx, ACL: tiger_acl.NewACL()}
	adapter.ResourceNodesKind = ResourceNodesKind
	adapter.RoleNodesKind = RoleNodesKind
	adapter.RulesKind = RulesKind
	return adapter
}
func NewDatastoreAdapterWithACL(ctx context.Context, ACL *tiger_acl.ACL) *DatastoreAdapter {
	adapter := NewDatastoreAdapter(ctx)
	adapter.ACL = ACL
	return adapter
}
func (adapter DatastoreAdapter) Save() error { return nil }
func (adapter DatastoreAdapter) Load() error {
	roleTreeRepository := datastore.NewDefaultRepository(adapter.ctx, adapter.RoleNodesKind)
	resourceTreeRepository := datastore.NewDefaultRepository(adapter.ctx, adapter.ResourceNodesKind)
	ruleRepository := datastore.NewDefaultRepository(adapter.ctx, adapter.RulesKind)
	// Roles
	roleNodes := []*RoleNode{}
	if err := roleTreeRepository.FindAll(&roleNodes); err != nil {
		return err
	}
	for _, node := range roleNodes {
		var parentRole tiger_acl.Role
		if node.GetParentRoleID() != "" {
			parentRole = tiger_acl.NewRole(node.GetParentRoleID())
		}
		adapter.ACL.AddRole(node, parentRole)
	}
	// Resources
	resourceNodes := []*ResourceNode{}
	if err := resourceTreeRepository.FindAll(&resourceNodes); err != nil {
		return err
	}
	for _, node := range resourceNodes {
		var parentResource tiger_acl.Resource
		if node.GetParentResourceID() != "" {
			parentResource = tiger_acl.NewResource(node.GetParentResourceID())
		}
		adapter.ACL.AddResource(node, parentResource)
	}
	// Rules
	rules := []*Rule{}
	if err := ruleRepository.FindBy(datastore.Query{Order: []string{"-Created"}}, &rules); err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.Type == tiger_acl.Allow {
			var (
				role     tiger_acl.Role
				resource tiger_acl.Resource
			)
			if rule.RoleID != "" {
				role = tiger_acl.NewRole(rule.RoleID)
			}
			if rule.ResourceID != "" {
				resource = tiger_acl.NewResource(rule.ResourceID)
			}
			if rule.AllPrivileges {
				adapter.ACL.Allow(role, resource)
			} else {
				adapter.ACL.Allow(role, resource, rule.Privilege)
			}
		}
	}
	return nil
}
