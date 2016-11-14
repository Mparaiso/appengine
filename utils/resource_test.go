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

package utils_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/Mparaiso/appengine/utils"
	"github.com/Mparaiso/go-tiger/test"
	"github.com/Mparaiso/go-tiger/validator"
	"google.golang.org/appengine/aetest"
)

type TestUser struct {
	Username,
	Email,
	PlainPassword string
	EncryptedPassword string `json:"-"`
	ID                int64
}

// GetUsername returns a string
func (testUser TestUser) GetUsername() string {
	return testUser.Username
}

// SetUsername sets *TestUser.testUser and returns *TestUser
func (testUser *TestUser) SetUsername(Username string) *TestUser {
	testUser.Username = Username
	return testUser
}

// GetID returns a int64
func (testUser TestUser) GetID() int64 {
	return testUser.ID
}

// SetID sets *TestUser.testUser and returns *TestUser
func (testUser *TestUser) SetID(ID int64) {
	testUser.ID = ID
}

func ValidateUser(ctx context.Context, r *http.Request, entity utils.Entity) error {
	user, ok := entity.(*TestUser)
	if !ok {
		return fmt.Errorf("Entity is not a TestUser")
	}
	errors := validator.NewValidationError()
	validator.EmailValidator("Email", user.Email, errors)
	return errors.ReturnNilOrErrors()
}

func TestResource(t *testing.T) {

	instance, err := aetest.NewInstance(nil)

	test.Fatal(t, err, nil)
	defer instance.Close()
	SubTestResourcePost(t, instance)

}

func SubTestResourcePost(t *testing.T, instance aetest.Instance) {
	LogFunc(t)
	resource := utils.NewResource(&TestUser{}, "users")
	resource.SetValidator(ValidateUser)
	buffer := new(bytes.Buffer)
	user := &TestUser{Username: "johndoe", Email: "johndoe@example.com", PlainPassword: "password"}
	err := json.NewEncoder(buffer).Encode(user)
	test.Fatal(t, err, nil)
	request, err := instance.NewRequest("POST", "/", buffer)
	test.Fatal(t, err, nil)
	response := httptest.NewRecorder()
	resource.Post(response, request)
	test.Fatal(t, response.Code, http.StatusCreated)
	message := &utils.CreatedMessage{}
	err = json.NewDecoder(response.Body).Decode(message)
	test.Fatal(t, err, nil)
	test.Fatal(t, message.ID != 0, true)
	SubTestResourcePost400(t, instance, resource)
	SubTestEndPointGet(t, instance, message.ID, resource)
	SubTestEndPointIndex(t, instance, resource)
	SubTestEndPointPut(t, instance, resource, message.ID)
	SubTestEndPointDelete(t, instance, resource, message.ID)
}

// Given an resource
// When "/" POST is required with an invalid entity
// it responds with 400
// it responds with an error message
func SubTestResourcePost400(t *testing.T, instance aetest.Instance, resource *utils.Resource) {
	buffer := new(bytes.Buffer)
	user := &TestUser{Username: "johndoe", PlainPassword: "password"}
	err := json.NewEncoder(buffer).Encode(user)
	test.Fatal(t, err, nil)
	request, err := instance.NewRequest("POST", "/", buffer)
	test.Fatal(t, err, nil)
	response := httptest.NewRecorder()
	resource.Post(response, request)
	test.Fatal(t, response.Code, http.StatusBadRequest)
	message := &struct{ Errors struct{ Email []string } }{}
	err = json.NewDecoder(response.Body).Decode(message)
	test.Fatal(t, err, nil)
	t.Log(err)
	test.Fatal(t, len(message.Errors.Email), 1, "Email errors should have 1 element")
}

func SubTestEndPointGet(t *testing.T, instance aetest.Instance, id int64, resource *utils.Resource) {
	LogFunc(t)
	request, err := instance.NewRequest("GET", fmt.Sprintf("/?:users=%d", id), nil)
	test.Fatal(t, err, nil)
	response := httptest.NewRecorder()
	resource.Get(response, request)
	test.Fatal(t, response.Code, http.StatusOK)
	user := &TestUser{}
	json.NewDecoder(response.Body).Decode(user)
	test.Error(t, user.ID, id)
	test.Error(t, user.Username, "johndoe")
}

func SubTestEndPointIndex(t *testing.T, instance aetest.Instance, resource *utils.Resource) {
	LogFunc(t)
	request, err := instance.NewRequest("GET", "/", nil)
	test.Fatal(t, err, nil)
	response := httptest.NewRecorder()
	resource.Index(response, request)
	test.Fatal(t, response.Code, 200)
	message := []*TestUser{}
	err = json.NewDecoder(response.Body).Decode(&message)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(message), 1, "[]*TestUser should have 1 element")
}

func SubTestEndPointPut(t *testing.T, instance aetest.Instance, resource *utils.Resource, ID int64) {
	body := new(bytes.Buffer)
	user := &TestUser{Username: "jackdoe", Email: "jackdoe@example.com"}
	err := json.NewEncoder(body).Encode(user)
	test.Fatal(t, err, nil)
	request, err := instance.NewRequest("PUT", fmt.Sprintf("/?:users=%d", ID), body)
	response := httptest.NewRecorder()
	resource.Put(response, request)
	test.Fatal(t, response.Code, 200, "PUT response should be 200")

	request, err = instance.NewRequest("GET", fmt.Sprintf("/?:users=%d", ID), nil)
	test.Fatal(t, err, nil)
	response = httptest.NewRecorder()
	resource.Get(response, request)
	test.Fatal(t, response.Code, 200)
	message := &TestUser{}
	err = json.NewDecoder(response.Body).Decode(message)
	test.Fatal(t, response.Code, 200)
	test.Fatal(t, message.Username, user.Username)
}

// Given an resource
// When Delete is requested with an ID
// It should respond with 200
// When Get is requested with the same ID
// It shouls respond with 404
func SubTestEndPointDelete(t *testing.T, instance aetest.Instance, resource *utils.Resource, ID int64) {
	response := httptest.NewRecorder()
	request, err := instance.NewRequest("DELETE", fmt.Sprintf("/?:users=%d", ID), nil)
	test.Fatal(t, err, nil)
	resource.Delete(response, request)
	test.Fatal(t, response.Code, 200)
	response = httptest.NewRecorder()
	request, err = instance.NewRequest("GET", fmt.Sprintf("/?:users=%d", ID), nil)
	test.Fatal(t, err, nil)
	resource.Get(response, request)
	test.Fatal(t, response.Code, 404, "requests resource should not exist")

}

// LogFunc the name of the test function
func LogFunc(t *testing.T) {
	// ptr, _, _, ok := runtime.Caller(1)
	// if ok {
	// 	t.Log(runtime.FuncForPC(ptr).Name(), "\n")
	// }
}
