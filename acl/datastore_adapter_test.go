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

package acl_test

import (
	"testing"

	"github.com/Mparaiso/appengine/acl"
	appengine_datastore "github.com/Mparaiso/appengine/datastore"
	tiger_acl "github.com/Mparaiso/go-tiger/acl"
	"github.com/Mparaiso/go-tiger/test"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestDatastoreAdapter_Load(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	test.Fatal(t, err, nil)
	defer done()
	parentKey := datastore.NewKey(ctx, "parent_keys", "parent", 0, nil)
	key, err := datastore.Put(ctx, parentKey, &struct{ Parent string }{"parent"})
	test.Fatal(t, err, nil)

	test.Fatal(t, datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		roleRepository := appengine_datastore.NewDefaultRepository(ctx, acl.RoleNodesKind)
		roleRepository.SetParentKey(key)
		resourceRepository := appengine_datastore.NewDefaultRepository(ctx, acl.ResourceNodesKind)
		resourceRepository.SetParentKey(key)
		ruleRepository := appengine_datastore.NewDefaultRepository(ctx, acl.RulesKind)
		ruleRepository.SetParentKey(key)
		test.Fatal(t, roleRepository.Create(&acl.RoleNode{RoleID: "guest"}), nil)
		test.Fatal(t, roleRepository.Create(&acl.RoleNode{RoleID: "staff", ParentRoleID: "guest"}), nil)
		test.Fatal(t, roleRepository.Create(&acl.RoleNode{RoleID: "administrator", ParentRoleID: "staff"}), nil)
		test.Fatal(t, resourceRepository.Create(&acl.ResourceNode{ResourceID: "article"}), nil)
		test.Fatal(t, resourceRepository.Create(&acl.ResourceNode{ResourceID: "page"}), nil)
		test.Fatal(t, ruleRepository.Create(&acl.Rule{Type: tiger_acl.Allow, RoleID: "guest", Privilege: "read", ResourceID: "article"}), nil)
		test.Fatal(t, ruleRepository.Create(&acl.Rule{Type: tiger_acl.Allow, RoleID: "staff", Privilege: "create", ResourceID: "article"}), nil)
		test.Fatal(t, ruleRepository.Create(&acl.Rule{Type: tiger_acl.Allow, RoleID: "staff", Privilege: "update", ResourceID: "article"}), nil)
		test.Fatal(t, ruleRepository.Create(&acl.Rule{Type: tiger_acl.Allow, RoleID: "administrator", Privilege: "delete"}), nil)
		var (
			roles     = []*acl.RoleNode{}
			resources = []*acl.ResourceNode{}
			rules     = []*acl.Rule{}
		)
		test.Fatal(t, roleRepository.FindAll(&roles), nil)
		test.Fatal(t, resourceRepository.FindAll(&resources), nil)
		test.Fatal(t, ruleRepository.FindAll(&rules), nil)
		t.Logf("%+v %+v %+v", roles, resources, rules)
		return nil
	}, nil), nil)

	adapter := acl.NewDatastoreAdapter(ctx)
	test.Fatal(t, adapter.Load(), nil)
	t.Logf("%+v %+v %+v", adapter.ACL.ResourceTree, adapter.ACL.RoleTree, adapter.ACL.Rules)
	test.Error(t, adapter.ACL.IsAllowed(tiger_acl.NewRole("guest"), tiger_acl.NewResource("page")), false)

}
