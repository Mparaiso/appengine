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
	Username,
	PlainPassword,
	EncryptedPassword string `json:"-"`
	ID int64
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
	SubTestEndpointPost(t, instance)

}

func SubTestEndpointPost(t *testing.T, instance aetest.Instance) {
	LogFunc(t)
	endpoints := utils.NewEndPoint(&TestUser{}, "users")
	buffer := new(bytes.Buffer)
	user := &TestUser{Username: "johndoe", PlainPassword: "password"}
	err := json.NewEncoder(buffer).Encode(user)
	test.Fatal(t, err, nil)
	request, err := instance.NewRequest("POST", "/", buffer)
	test.Fatal(t, err, nil)
	response := httptest.NewRecorder()
	endpoints.Post(response, request)
	test.Fatal(t, response.Code, http.StatusCreated)
	message := &utils.CreatedMessage{}
	err = json.NewDecoder(response.Body).Decode(message)
	test.Fatal(t, err, nil)
	test.Fatal(t, message.ID != 0, true)
	SubTestEndPointGet(t, instance, message.ID, endpoints)
	SubTestEndPointIndex(t, instance, endpoints)
}

func SubTestEndPointGet(t *testing.T, instance aetest.Instance, id int64, endpoints *utils.EndPoint) {
	LogFunc(t)
	request, err := instance.NewRequest("GET", fmt.Sprintf("/?:users=%d", id), nil)
	test.Fatal(t, err, nil)
	response := httptest.NewRecorder()
	endpoints.Get(response, request)
	test.Fatal(t, response.Code, http.StatusOK)
	user := &TestUser{}
	json.NewDecoder(response.Body).Decode(user)
	t.Logf("%+v", user)
	test.Error(t, user.ID, id)
	test.Error(t, user.Username, "johndoe")
}

func SubTestEndPointIndex(t *testing.T, instance aetest.Instance, endpoints *utils.EndPoint) {
	LogFunc(t)
}

// LogFunc the name of the test function
func LogFunc(t *testing.T) {
	ptr, _, _, ok := runtime.Caller(1)
	if ok {
		t.Log("=== ", runtime.FuncForPC(ptr).Name(), " ===", "\n")
	}
}
