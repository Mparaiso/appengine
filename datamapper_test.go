package datamapper_test

import (
	"testing"

	. "github.com/mparaiso/datamapper"
)

func TestRegister(t *testing.T) {
	connection := initializeConnection(t)
	defer connection.Close()
	dm := NewDataMapper(connection)
	if err := dm.Register(new(NotEntity)); err == nil {
		t.Fatal("Should return an error since NotEntity is not a valid entity")
	}
	dm = NewDataMapper(connection)
	if err := dm.Register(new(User)); err != nil {
		t.Fatalf("error should be nil while registering valid MetadataProvider, got %s.", err)
	}
}

func TestGetRepository(t *testing.T) {
	connection := initializeConnection(t)
	defer connection.Close()
	dm := NewDataMapper(connection)
	err := dm.Register(new(User))
	if err != nil {
		t.Fatal(err)
	}
	repository, err := dm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	if repository == nil {
		t.Fatalf("repository of entities %#v should not be nil.", new(User))
	}
}

func TestGetRepositoryByTableName(t *testing.T) {
	connection := initializeConnection(t)
	defer connection.Close()
	dm := NewDataMapper(connection)
	err := dm.Register(new(User))
	if err != nil {
		t.Fatal(err)
	}
	repository, err := dm.GetRepositoryByTableName("users")
	if err != nil {
		t.Fatal(err)
	}
	if repository == nil {
		t.Fatal("repository of for table 'users' should not be nil.")
	}
}
