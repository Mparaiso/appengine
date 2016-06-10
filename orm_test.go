package orm_test

import (
	"database/sql"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/mparaiso/go-orm"
	"github.com/rubenv/sql-migrate"
)

func TestORMPersist(t *testing.T) {
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(User), new(Article))
	if err != nil {
		t.Fatal(err)
	}
	user := &User{Name: "John", Email: "john@acme.com"}
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	user.Name = "Jack"
	user.AddArticles([]*Article{{Title: "First Article Title"}, {Title: "Second Article Title"}}...)
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	ThenTestRepositoryFind(orm, t, user.ID)
}

func ThenTestRepositoryFind(orm *ORM, t *testing.T, userID int64) {
	userRepository, err := orm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	user := new(User)
	err = userRepository.Find(userID, user)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(user.Articles); l != 2 {
		t.Fatalf("Articles length should be 2, got %d", l)
	}
}

func TestRepositoryFindBy(t *testing.T) {
	user := &User{Name: "John Doe", Email: "john.doe@acme.com"}
	articles := []*Article{{Title: "First Article"}, {Title: "Second Article"}}
	orm := NewORM(GetConnection(t))
	orm.Register(new(User), new(Article))
	err := orm.Persist(user).Flush()
	if err != nil {
		t.Fatal(err)
	}
	user.AddArticles(articles...)
	err = orm.Persist(user).Flush()
	if err != nil {
		t.Fatal(err)
	}
	userRepository, err := orm.GetRepository(user)
	if err != nil {
		t.Fatal(err)
	}
	result := []*User{}
	err = userRepository.FindBy(Query{Where: []string{"ID", "=", "?"}, Params: []interface{}{user.ID}}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(user.Articles); l != 2 {
		t.Fatalf("length should b 2, got %d", l)
	}
	ThenTestRepositoryAll(orm, t)
}

func ThenTestRepositoryAll(orm *ORM, t *testing.T) {
	userRepository, err := orm.GetRepository(new(User))
	if err != nil {
		t.Fatal(err)
	}
	result := []*User{}
	err = userRepository.All(&result)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 1 {
		t.Fatalf("Len Should be 1, got %d", l)
	}

	ThenTestRepositoryFindOneBy(userRepository, t, result[0])
}

func ThenTestRepositoryFindOneBy(repository *Repository, t *testing.T, user *User) {
	result := new(User)
	err := repository.FindOneBy(Query{Where: []string{"ID", "=", "?"}, Params: []interface{}{user.ID}}, result)
	if err != nil {
		t.Fatal(err)
	}
	if id := result.ID; id != user.ID {
		t.Fatalf("The ID of the user should be %d, got %d", user.ID, result.ID)
	}

}

func TestORMDestroy(t *testing.T) {
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(User), new(Article))
	if err != nil {
		t.Fatal(err)
	}
	user := &User{Name: "John", Email: "john@acme.com"}
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	orm.Remove(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Destroy entity that has a one to many relationship with cascade Remove of owned entities")
	user = GetUserFixture()[1]
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	userID := user.ID
	articles := GetArticleFixture()
	user.AddArticles(articles[:2]...)
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	orm.Remove(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	articleRepository, err := orm.GetRepository(new(Article))
	if err != nil {
		t.Fatal(err)
	}
	result := []*Article{}
	articleRepository.FindBy(Query{Where: []string{"AuthorID", "=", "?"}, Params: array{userID}}, &result)
	if l := len(result); l != 0 {
		t.Fatalf("Length should be 0, got %d", l)
	}
}

func TestRegressionBug(t *testing.T) {
	user := new(User)
	orm := NewORM(GetConnection(t))
	err := orm.Register(new(Article), user, new(UserInfo))
	userRepository, err := orm.GetRepository(user)
	query, values, err := Query{
		Set:  user.ProvideMetadata().BuildFieldValueMap(user),
		Type: INSERT,
	}.BuildQuery(userRepository)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(query, values)
}
func TestRegressionBug2(t *testing.T) {

	orm := NewORM(GetConnection(t))
	err := orm.Register(new(Article), new(User), new(UserInfo))
	if err != nil {
		t.Fatal(err)
	}
	userFixtures := GetUserFixture()
	user := userFixtures[0]
	orm.Persist(user)
	err = orm.Flush()
	if err != nil {
		t.Fatal(err)
	}
	fetchedUser := new(User)
	userRepository, err := orm.GetRepository(user)
	if err != nil {
		t.Fatal(err)
	}
	err = userRepository.Find(user.ID, fetchedUser)
	if err != nil {
		t.Fatal(err)
	}
}

type array []interface{}

func GetUserFixture() []*User {
	return []*User{
		{Name: "john doe", Email: "john.doe@acme.com"},
		{Name: "jane doe", Email: "jane.doe@acme.com"},
		{Name: "bill doe", Email: "bill.doe@acme.com"},
		{Name: "suzy doe", Email: "suzy.doe@acme.com"},
	}
}

func GetArticleFixture() []*Article {
	return []*Article{
		{Title: "First Article Content", Content: "First Article Content"},
		{Title: "Second Article Content", Content: "Second Article Content"},
	}
}

func GetConnection(t *testing.T) *Connection {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: "testdata/migrations/development.sqlite3",
	}
	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}

	return NewConnectionWithOptions("sqlite3", db, &ConnectionOptions{Logger: t})
}

// Article is an article
type Article struct {
	ID       int64
	Title    string
	Content  string
	Created  time.Time
	Updated  time.Time
	Author   *User
	AuthorID int64
}

func (Article) ProvideMetadata() Metadata {
	return Metadata{
		Entity: "Article",
		Table:  Table{Name: "articles"},
		Columns: []Column{
			{ID: true, StructField: "ID"},
			{StructField: "Title"},
			{StructField: "Content"},
			{StructField: "Created"},
			{StructField: "Updated"},
			{StructField: "AuthorID", Name: "author_id"},
		},
	}
}

func (a *Article) BeforeCreate() (err error) {
	a.Created = time.Now()
	a.Updated = time.Now()
	return
}

func (a *Article) BeforeUpdate() (err error) {
	a.Updated = time.Now()
	return
}

// User is a user
type User struct {
	ID             int64
	Name           string
	Email          string
	Created        time.Time
	Updated        time.Time
	PasswordDigest string
	Articles       []*Article
	*UserInfo
}

func (User) ProvideMetadata() Metadata {
	return Metadata{
		Entity: "User",
		Table:  Table{Name: "users"},
		Columns: []Column{
			{ID: true, StructField: "ID"},
			{StructField: "Email"},
			{StructField: "Name"},
			{StructField: "Created"},
			{StructField: "Updated"},
			{StructField: "PasswordDigest", Name: "password_digest"},
		},
		Relations: []Relation{
			{
				StructField:  "Articles",
				Type:         OneToMany,
				TargetEntity: "Article",
				IndexBy:      "AuthorID",
				MappedBy:     "Author",
				Fetch:        Eager,
				Cascade:      Persist | Remove,
			},
		},
	}
}

func (user *User) BeforeCreate() (err error) {
	user.Created = time.Now()
	user.Updated = time.Now()
	return
}

func (user *User) BeforeUpdate() (err error) {
	user.Updated = time.Now()
	return
}

func (user *User) AddArticles(articles ...*Article) {
	if user.Articles == nil {
		user.Articles = []*Article{}
	}
	for _, article := range articles {
		user.Articles = append(user.Articles, article)
		article.AuthorID = user.ID
	}
}

// see http://stackoverflow.com/questions/23259586/bcrypt-password-hashing-in-golang-compatible-with-node-js
func (user *User) GenerateSecurePassword(password string) error {
	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(passwordDigest)
	return nil
}

// Authenticate return an error if the password and PasswordDigest do not match
func (user User) Authenticate(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))
}

// UserInfo complements the user entity
type UserInfo struct {
	ID            int64
	NiceName      string
	URL           string
	Registered    time.Time
	ActivationKey string
	Status        int8
	DisplayName   string
	UserID        int64
	User          *User
}

func (UserInfo) ProvideMetadata() Metadata {
	return Metadata{
		Table:  Table{Name: "user_infos"},
		Entity: "UserInfo",
		Columns: []Column{
			{StructField: "ID", ID: true},
			{StructField: "NiceName"},
			{StructField: "Registered"},
			{StructField: "ActivationKey", Name: "activation_key"},
			{StructField: "Status"},
			{StructField: "DisplayName", Name: "display_name"},
			{StructField: "UserID", Name: "user_id"},
		},
	}
}

// NotEntity is not a valid entity
type NotEntity struct{}
