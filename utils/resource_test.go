package utils_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/Mparaiso/appengine/utils"
	"github.com/Mparaiso/go-tiger/test"
	"google.golang.org/appengine/aetest"
)

type TestUser struct {
	Username          string
	PlainPassword     string
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

func TestEndPoint(t *testing.T) {

	instance, err := aetest.NewInstance(nil)

	test.Fatal(t, err, nil)
	defer instance.Close()
	SubTestResourcePost(t, instance)

}

func SubTestResourcePost(t *testing.T, instance aetest.Instance) {
	LogFunc(t)
	resource := utils.NewResource(&TestUser{}, "users")
	buffer := new(bytes.Buffer)
	user := &TestUser{Username: "johndoe", PlainPassword: "password"}
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
	SubTestEndPointGet(t, instance, message.ID, resource)
	SubTestEndPointIndex(t, instance, resource)
	SubTestEndPointPut(t, instance, resource, message.ID)
	SubTestEndPointDelete(t, instance, resource, message.ID)
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
	t.Logf("%+v", user)
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
	test.Error(t, len(message), 1)
}

func SubTestEndPointPut(t *testing.T, instance aetest.Instance, resource *utils.Resource, ID int64) {
	LogFunc(t)
	body := new(bytes.Buffer)
	user := &TestUser{Username: "jackdoe"}
	err := json.NewEncoder(body).Encode(user)
	test.Fatal(t, err, nil)
	request, err := instance.NewRequest("PUT", fmt.Sprintf("/?:users=%d", ID), body)
	response := httptest.NewRecorder()
	resource.Put(response, request)
	test.Fatal(t, response.Body, 200)
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
func SubTestEndPointDelete(t *testing.T, instance aetest.Instance, resource *utils.Resource, ID int64) {
	response := httptest.NewRecorder()
	request, err := instance.NewRequest("DELETE", fmt.Sprintf("/?:users=%d", ID), nil)
	test.Fatal(t, err, nil)
	resource.Delete(response, request)
	test.Fatal(t, response.Code, 200)
	response = httptest.NewRecorder()
	request, err = instance.NewRequest("GET", fmt.Sprintf("/?:users=%d", ID), nil)
	test.Fatal(t, err, nil)
	resource.Delete(response, request)
	test.Fatal(t, response.Code, 404)

}

// LogFunc the name of the test function
func LogFunc(t *testing.T) {
	ptr, _, _, ok := runtime.Caller(1)
	if ok {
		t.Log(runtime.FuncForPC(ptr).Name(), "\n")
	}
}
